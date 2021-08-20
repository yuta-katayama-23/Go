package main

import (
	"context"
	"fmt"
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
	// Slackからの実行用
	Modules      []string `json:"Modules"`
	SourceBranch string   `json:"SourceBranch"`
	TargetEnv    string   `json:"TargetEnv"`

	// Cloud Watch Eventsからの実行用
	Detail Detail `json:"detail"`
}

type Detail struct {
	Event           string   `json:"event"`
	RepositoryNames []string `json:"repositoryNames"`
	SourceReference string   `json:"sourceReference"`
	Title           string   `json:"title"`
}

func HandleRequest(ctx context.Context, event Event) (string, error) {
	configor.Load(&Config, "config.yaml")

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

	return fmt.Sprintf("Module %s", event.Modules), nil
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
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
