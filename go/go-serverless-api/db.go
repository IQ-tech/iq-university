package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/satori/go.uuid"
)

var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("sa-east-1"))
const tableName = "go-serverless-api"

// GetTodo retrieves one Todo from the DB based on its ID
func GetTodo(uuid string) (Todo, error) {

	// Prepares the input to retrieve the item with the given ID
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(uuid),
			},
		},
	}

	// Retrieves the item
	result, err := db.GetItem(input)
	if err != nil {
		return Todo{}, err
	}
	if result.Item == nil {
		return Todo{}, nil
	}

	// Unmarshals the object retrieved into a domain struct
	var todo Todo
	err = dynamodbattribute.UnmarshalMap(result.Item, &todo)
	if err != nil {
		return Todo{}, err
	}

	return todo, nil
}

// GetTodos retrieves all the Todos from the DB
func GetTodos() ([]Todo, error) {

	// Prepares the input to scan the whole table
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}
	result, err := db.Scan(input)
	if err != nil {
		return []Todo{}, err
	}
	if len(result.Items) == 0 {
		return []Todo{}, nil
	}

	// Unmarshals the array retrieved into a domain struct's slice
	var todos []Todo
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &todos)
	if err != nil {
		return []Todo{}, err
	}

	return todos, nil
}

// CreateTodo inserts a new Todo item to the table.
func CreateTodo(todo Todo) error {

	// Generates a new random ID
	uuid, err := uuid.NewV4()
	if err != nil {
		return err
	}

	// Creates the item that's going to be inserted
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(fmt.Sprintf("%v", uuid)),
			},
			"title": {
				S: aws.String(todo.Title),
			},
			"description": {
				S: aws.String(todo.Description),
			},
		},
	}

	_, err = db.PutItem(input)
	return err
}
