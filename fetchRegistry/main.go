package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/liuminhaw/sportcenter-booking/registry"
)

func fetchRegistry(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	queryStrings := event.QueryStringParameters
	if _, ok := queryStrings["id"]; !ok {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusForbidden,
			Body:       "Request forbidden\n",
		}, nil
	}

	sess, err := session.NewSession()
	if err != nil {
		fmt.Println(err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, err
	}
	svc := s3.New(sess)

	registryId := queryStrings["id"]
	s3Registry := registry.Registry{
		Bucket:   os.Getenv("S3Bucket"),
		Dirname:  "registry",
		Filename: registryId,
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(s3Registry.Bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s3Registry.Dirname, s3Registry.Filename)),
	}

	fmt.Println("S3 Bucket: ", s3Registry.Bucket)
	fmt.Println("S3 Key: ", fmt.Sprintf("%s/%s", s3Registry.Dirname, s3Registry.Filename))

	_, err = svc.HeadObject(input)
	if err == nil {
		err := s3Registry.FetchRegistryFile(sess, os.Getenv("secretKey"))
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       err.Error(),
			}, nil
		}
	} else {
		fmt.Println(err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusForbidden,
			Body:       "Forbidden\n",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(s3Registry.Content),
	}, nil
}

func main() {
	lambda.Start(fetchRegistry)
}
