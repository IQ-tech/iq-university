# go-serverless-api

Deploys a Golang Servlerss API to AWS using Lambda, DynamoDB and API Gateway.

## Table of Contents

  * [Prerequisites](#prerequisites)
  * [Setting up the environment](#setting-up-the-environment)
    + [Creating the Lambda function](#creating-the-lambda-function)
    + [Creating a new DynamoDB table](#creating-a-new-dynamodb-table)
    + [Creating a new API Gateway](#creating-a-new-api-gateway)
  * [Coding the application](#coding-the-application)
  * [Deploying to AWS](#deploying-to-aws)


## Prerequisites

* [Golang](https://golang.org/dl/)
* [AWS account](https://aws.amazon.com/)

## Setting up the environment

### Creating the Lambda function

* Go to the [Lambda service](https://console.aws.amazon.com/lambda/home);
* Click in the **Create a function** button, give your function a name (we used `go-serverless-api`) and choose **Go** as the Runtime;
* Create a new role with permission to manage your application: choose **Create new role from template**, give your role a name (we named it `microservice-role`) and then choose **Simple Microservice permissions** as the Policy template;
* In the new page, click in the select input with "*Select a test event...*" and select **Configure test events**;
* Choose **Create new test event**, name it as you want (we named it `Blank`) and leave its body empty (`{}`).

### Creating a new DynamoDB table

* Go to the [DynamoDB service] (https://console.aws.amazon.com/dynamodb/home);
* Click in the **Create table** button, give you table a name (we chose `go-serverless-api`) and define its primary key. For out application, we're going to use `id` as our primary key.

### Creating a new API Gateway

* Go to the [API Gateway](https://console.aws.amazon.com/apigateway/home) service and click in the **Get Started** button;
* Choose **New API** in the radio buttons and give you API a name (one more time we'll name it `go-serverless-api`).
* In the **Resources** page, click in the resource that was already created for you (`/`), click in **Actions** and select **Create Resource**.
* Give your resource a name and specify it's name (we'll name it **Todos** and specify its path as `todos`).
* Now click in your `/todos` resource and repeat the process: click in the **Actions** button, give it a name (we'll use **SingleTodo**) and a path. Since we're going to use this path to retrieve a Todo by its ID, we're going to specify a path parameter, so our resource path is going to be `/todos/{id}`.
* Now, let's create the methods: click in the `/todos` resource, click in the **Actions** button, select **Create Method** and choose `GET`. in the method setup, choose **Lambda Function** as the Integration type, check the **Use Lambda Proxy integration** checkbox, choose the Region where your Lambda is and enter its name (`go-serverless-api` in our case). Repeat this process to create a `POST` method to `/todos` and a `GET` method to `/todos/{id}`.

## Coding the application

Lambda functions on Go have, by default, the following structure:

```golang
func main() {
  lambda.Start(handler)
}
```

Since our application is going to deal with 3 different routes, we need a router to deal with those requests, so our `main.go` file is going to be like this:

```golang
// main.go

package main

import (
  "net/http"
  "regexp"

  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
)

// handleCreateTodo

// handleGetTodo

// handleGetTodos

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

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
```

You may be asking why we're using the `events` package from AWS, right? Basically, `events` will allow us to retrieve the request from API Gateway and to send our response in such a way it can be read properly.
Passing `req events.APIGatewayProxyRequest` as a parameter to our `router` function, we'll have access to all the parameters from the request such as `Path` and `HTTPMethod` that we're going to use to call the correct handler.

That said, we have:

* `GET /todos`: retrieves the list containing all the Todos from the database
* `POST /todos`: inserts a new Todo into the database
* `GET /todos/{id}`: retrieves a Todo from the database given its ID.

Before creating the handlers, let's create our Todo and the functions that will deal with our database. First of all, create a new file called `domain.go` and define the Todo's struct:

```golang
// domain.go

package main

type Todo struct {
  ID          string `json:"id,omitempty"`
  Title       string `json:"title"`
  Description string `json:"description"`
  Date        string `json:"date"`
}
```

In order to make it simple, we'll use only `string` attributes: an ID (our primary key), a title and a description.

Now, create a `db.go` file and insert some lines of code that are going to allow you to connect your application to your DynamoDB:

```golang
// db.go

package main

import (
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/dynamodb"
)
const AWS_REGION = "sa-east-1"
const TABLE_NAME = "go-serverless-api"

var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion(AWS_REGION))
```

Now that we have the `db` instance, we can start to "query" our database. We're going to use three operations:

* **GetItem:** retrieves an item with a certain primary key;
* **PutItem:** inserts a new item into the table;
* **Scan:** scans the whole table listing all the items.

All the DynamoDB operations follow a pattern: generate an input, send that input to DynamoDB, unmarshal the result (if needed) and then return the result of the operation.

In order to unmarshal the response, we can use `UnmarshalListOfMaps()` and `UnmarshalMap`, both from the AWS package.

Here's `GetTodos` as an example (the other 2 operations can be found in this repository):

```golang
// db.go

func GetTodos() ([]Todo, error) {

  input := &dynamodb.ScanInput{
    TableName: aws.String(TABLE_NAME),
  }
  result, err := db.Scan(input)
  if err != nil {
    return []Todo{}, err
  }
  if len(result.Items) == 0 {
    return []Todo{}, nil
  }

  var todos []Todo
  err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &todos)
  if err != nil {
    return []Todo{}, err
  }

  return todos, nil
}
```

Now that we have all our DB operations set up, we can proceed to our main file and create our handlers. In the same way as our database operations, our handlers follow a very well defined pattern: handle your input/request (if needed to retrieve parameters, for example), call the database operation, handle the errors, convert your response to a string and return it. In the same way as our request, our response has to be encapsulated by a `APIGatewayProxyResponse` that requires a status code and the response's body (you have to encapsulate it even if it's an error!).

As an example, here's our handler responsible by the `GET /todos` route:

```golang
// main.go

func handleGetTodos(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

  todos, err := GetTodos()
  if err != nil {
    fmt.Println(err)
    return events.APIGatewayProxyResponse{
      StatusCode: http.StatusInternalServerError,
      Body:       http.StatusText(http.StatusInternalServerError),
    }, nil
  }

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
```

## Deploying to AWS

Now that our application is all set up, we're ready to deploy it to Amazon in a few steps. First, we're going to send our application to our Lambda function:

1. Compile your code using: 

```shell
# Compile the code in the main file
$ GOOS=linux go build -o main

# Put your executable into a zip file
$ zip main.zip main
```

2. Go to your Lambda function at your AWS Console.
3. In the **Function code** section, upload your `.zip` file and change your Handler to `main` (since it's the name of our main function).
4. Click **Save** at the top of the page.

Now we can test our API: go to the API Gateway we created earlier, select one of the routes (let's start with the `GET /todos` route) and click in the **TEST** icon. Since we don't use any parameters in that request, just click the **TEST** button. You'll see that our response is an empty array (as expected!).

Let's create some items in our database... Select the `POST /todos` route, click in the **TEST** icon and fill the Request Body with:

```json
{
    "title": "That's a new To Do!",
    "description": "Testing the DB"
}
```

Execute the test and you may receive the `Created` response. Nice! Now execute the `GET /todos` request one more time and you'll see that our ToDo was really created in our database.

Since it's working fine, let's deploy our API. Click in the **Actions** button, select **Deploy API**, create a new stage (we'll name it as `go-serverless-api`) and you'll get the **Invoke URL** that we're going to use to call our API. Send a request to `<invoke-url>/todos` and you'll be able to retrieve all the itmes from your database!