package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const respBodyContent = `Username: %s
Password: %s
Reserve date: %v
Reserve court: %s
Reserve time: %s
`

type reservation struct {
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	ReserveDate  time.Time `json:"reserveDate"`
	ReserveCourt string    `json:"reserveCourt"`
	ReserveTime  string    `json:"reserveTime"`
}

func reservationRegistry(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var data reservation

	if err := json.Unmarshal([]byte(event.Body), &data); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, err
	}

	respBody := fmt.Sprintf(respBodyContent,
		data.Username, data.Password, data.ReserveDate, data.ReserveCourt, data.ReserveTime)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       respBody,
	}, nil
}

func main() {
	lambda.Start(reservationRegistry)
}
