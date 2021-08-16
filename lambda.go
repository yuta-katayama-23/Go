package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/configor"
)

var Config = struct {
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

func main() {
	lambda.Start(HandleRequest)
}
