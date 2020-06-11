package main

import (
	"backend/common"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	bestFruit, ok := request.QueryStringParameters["best-fruit"]
	if ok {
		if bestFruit == common.BestFruit {
			return events.APIGatewayProxyResponse{Headers: request.Headers, Body: `{"best-fruit": "how did you know?"}`, StatusCode: 200}, nil
		}
		if bestFruit == common.BestFruit {
			return events.APIGatewayProxyResponse{Headers: request.Headers, Body: fmt.Sprintf(`{"best-fruit": "not %s"}`, bestFruit), StatusCode: 200}, nil
		}
	}

	b, err := json.Marshal(request.QueryStringParameters)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	return events.APIGatewayProxyResponse{Headers: request.Headers, Body: string(b), StatusCode: 200}, nil
}

func main() {

	lambda.Start(Handler)

}
