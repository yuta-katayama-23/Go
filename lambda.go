package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

type Event struct {
	Name   string `json:"name"`
	Test   string `json:"test"`
	Source string `json:"source"`
	Detail string `json:"detail"`
}

func HandleRequest(ctx context.Context, event Event) (string, error) {
	fmt.Println(event.Test)
	fmt.Println(event.Source)
	fmt.Println(event.Detail)
	return fmt.Sprintf("Hello %s!", event.Name), nil
}

func main() {
	lambda.Start(HandleRequest)
}
