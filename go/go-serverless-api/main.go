package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleGetTodo(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	
	// Retrieves the ID from the URL
	id := req.PathParameters["id"]

	// Fetches the requested Todo
	todo, err := GetTodo(id)
	if err != nil {
		fmt.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       http.StatusText(http.StatusInternalServerError),
		}, nil
	}

	// Marshals the struct so the API Gateway is able to proccess it
	js, err := json.Marshal(todo)
	if err != nil {
		fmt.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       http.StatusText(http.StatusInternalServerError),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func handleGetTodos(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	
	// Fetches all the Todos
	todos, err := GetTodos()
	if err != nil {
		fmt.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       http.StatusText(http.StatusInternalServerError),
		}, nil
	}

	// Marshals the struct so the API Gateway is able to proccess it
	js, err := json.Marshal(todos)
	if err != nil {
		fmt.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       http.StatusText(http.StatusInternalServerError),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func handleCreateTodo(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// Unmarshals the request's body
	var todo Todo
	err := json.Unmarshal([]byte(req.Body), &todo)
	if err != nil {
		fmt.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	// Inserts the Todo into the table
	err = CreateTodo(todo)
	if err != nil {
		fmt.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Body:       "Created",
	}, nil
}

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// Routes the application to the correct handler based in the request's path
	if req.HTTPMethod == "GET" {
		hasID, _ := regexp.MatchString("/todos/.+", req.Path)
		if hasID {
			return handleGetTodo(req)
		}

		if req.Path == "/todos" {
			return handleGetTodos(req)
		}
	}
	if req.HTTPMethod == "POST" {
		return handleCreateTodo(req)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusMethodNotAllowed,
		Body:       http.StatusText(http.StatusMethodNotAllowed),
	}, nil
}

func main() {
	lambda.Start(router)
}
