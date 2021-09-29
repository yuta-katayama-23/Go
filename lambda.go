package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/configor"
	"github.com/slack-go/slack"
)

var Config = struct {
	SlackSetting struct {
		ApiToken string `yaml:"ApiToken"`
		Channel  string `yaml:"Channel"`
	} `yaml:"SlackSetting"`

	BuildRepositories []struct {
		TargetEnv     string          `yaml:"TargetEnv"`
		BuildProjects []BuildProjects `yaml:"BuildProjects"`
	} `yaml:"BuildRepositories"`
}{}

type BuildProjects struct {
	ProjectName string `yaml:"ProjectName"`
	ModuleName  string `yaml:"ModuleName"`
}

type Event struct {
	Modules      []string `json:"Modules"`
	SourceBranch string   `json:"SourceBranch"`
	TargetEnv    string   `json:"TargetEnv"`
	IssueKey     string   `json:"issueKey"`

	Detail Detail `json:"detail"`
}

type Detail struct {
	Event           string   `json:"event"`
	RepositoryNames []string `json:"repositoryNames"`
	SourceReference string   `json:"sourceReference"`
	Title           string   `json:"title"`
}

type Issue struct {
	Status Status `json:"status"`
}

type Status struct {
	Id        int
	ProjectId int
	Name      string
}

func HandleRequest(ctx context.Context, event Event) (string, error) {
	configor.Load(&Config, "config.yaml")

	fmt.Println("Config", Config)

	checkResult := checkIssueStatus(event.IssueKey)
	if checkResult == "ok" {
		var buildProjects []BuildProjects

		for _, repository := range Config.BuildRepositories {
			if repository.TargetEnv == event.TargetEnv {
				buildProjects = repository.BuildProjects
			}
		}

		fmt.Println(buildProjects)

		for _, project := range buildProjects {
			if contains(event.Modules, project.ModuleName) {
				fmt.Println(project.ProjectName)
			}
		}

		res, err := slackNotification(event)
		fmt.Println("res", res)
		fmt.Println("err", err)

		updateResult := updateIssueStatus(event.IssueKey)
		if updateResult != "ok" {
			return fmt.Sprintf("Update Issue Status Error is %s.", updateResult), nil
		}

		return fmt.Sprintf("Build Sucess and Issue Status is Updated"), nil
	} else {
		return fmt.Sprintf("Error is %s.", checkResult), nil
	}
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}

func checkIssueStatus(issueKey string) string {
	reqUrl := "https://" + os.Getenv("BACKLOG_DOMEIN") + ".backlog.com/api/v2/issues/" + issueKey + "?apiKey=" + os.Getenv("BACKLOG_API_KEY")
	resp, err := http.Get(reqUrl)
	if err != nil {
		fmt.Println("Request Error:", err)
		return "Request Error:" + err.Error()
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Response Error:", resp.Status)
		return "Response Error:" + resp.Status
	}

	body, err := io.ReadAll(resp.Body)
	fmt.Println(string(body))

	var issue Issue
	json.Unmarshal(body, &issue)

	if issue.Status.Name != os.Getenv("BACKLOG_ISSUE_STATUS") {
		fmt.Println("Status Error:", issue.Status.Name)
		return "Status Error:" + issue.Status.Name
	}
	return "ok"
}

func updateIssueStatus(issueKey string) string {
	reqUrl := "https://" + os.Getenv("BACKLOG_DOMEIN") + ".backlog.com/api/v2/issues/" + issueKey + "?apiKey=" + os.Getenv("BACKLOG_API_KEY")
	values := url.Values{}
	values.Set("statusId", os.Getenv("BACKLOG_STATUS_ID"))

	req, err := http.NewRequest(
		"PATCH",
		reqUrl,
		strings.NewReader(values.Encode()),
	)

	if err != nil {
		fmt.Println("NewRequest Error:", err)
		return "NewRequest Error:" + err.Error()
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := new(http.Client)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Request Error:", err)
		return "Request Error:" + err.Error()
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Response Error:", resp.Status)
		return "Response Error:" + resp.Status
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))

	return "ok"
}

func slackNotification(event Event) (string, error) {
	client := slack.New(os.Getenv("SLACK_API_TOKEN"))

	headerText := slack.NewTextBlockObject("mrkdwn", "<!channel>\nこれより *"+event.TargetEnv+"環境* リリース作業を実施致します。\n完了後、通知致します。", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	complete := slack.NewTextBlockObject("mrkdwn", "*完了予定*\n15分", false, false)
	target := slack.NewTextBlockObject("mrkdwn", "*対象環境*\n"+event.TargetEnv, false, false)
	module := slack.NewTextBlockObject("mrkdwn", "*対象サービス*\n"+strings.Join(event.Modules, ", "), false, false)
	fieldSlice := make([]*slack.TextBlockObject, 0)
	fieldSlice = append(fieldSlice, complete)
	fieldSlice = append(fieldSlice, target)
	fieldSlice = append(fieldSlice, module)
	fieldsSection := slack.NewSectionBlock(nil, fieldSlice, nil)

	_, _, err := client.PostMessage(os.Getenv("SLACK_CHANNEL"), slack.MsgOptionBlocks(headerSection, fieldsSection))
	if err != nil {
		fmt.Printf("%s\n", err)
		return fmt.Sprintf("Fail slack notification\nError message is \"%s\"", err), nil
	}
	return "Success slack notification", nil
}

func main() {
	lambda.Start(HandleRequest)
}
