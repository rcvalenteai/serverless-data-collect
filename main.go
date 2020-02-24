package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
)

var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

type Response struct {
	Message string `json:"message"`
	Ok      bool   `json:"ok"`
}

func show(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var experiment Experiment
	err := json.Unmarshal([]byte(req.Body), &experiment)
	if err != nil {
		fmt.Printf("error unmarshalling json body")
	}
	insertedId := sendMongo(experiment)
	if err != nil {
		return serverError(err)
	}

	headers := map[string]string{"Access-Control-Allow-Origin": "*", "Access-Control-Allow-Headers": "Origin, X-Requested-With, Content-Type, Accept"}
	js, err := json.Marshal(getResponse(insertedId))
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
		Headers:    headers,
	}, nil
}

func getResponse(id string) Response {
	return Response{
		Message: fmt.Sprintf("Successfully added: %s", id),
		Ok:      true}
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
	errorLogger.Println(err.Error())

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

type Experiment struct {
	Charts []struct {
		ChartType string `json:"chartType"`
	} `json:"charts"`
	TrialsPerChart int `json:"trialsPerChart"`
	Image          struct {
		Groups [][]struct {
		} `json:"_groups"`
		Parents []struct {
		} `json:"_parents"`
	} `json:"image"`
	TrialCount int `json:"trialCount"`
	Trials     []struct {
		Chart struct {
			ChartType string `json:"chartType"`
		} `json:"chart"`
		RandomValues      []float64 `json:"randomValues"`
		TestIndices       []int     `json:"testIndices"`
		RealPercentage    float64   `json:"realPercentage"`
		GuessedPercentage int       `json:"guessedPercentage"`
	} `json:"trials"`
	Timestamp int64 `json:"timestamp"`
}

func sendMongo(experiment Experiment) string {
	password := os.Getenv("MONGO_PASS")
	url := "mongodb+srv://bigdatamanagement:" + password + "@mongosandbox-qzmsu.mongodb.net/test?retryWrites=true&w=majority"
	clientOptions := options.Client().ApplyURI(url)
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("projectfour").Collection("datavisthree")
	insertResult, err := collection.InsertOne(context.TODO(), experiment)
	if err != nil {
		log.Fatal(err)
	}
	newID := insertResult.InsertedID
	fmt.Println("Inserted a Single Document: ", newID)
	return newID.(primitive.ObjectID).String()
}

func main() {
	lambda.Start(show)
}
