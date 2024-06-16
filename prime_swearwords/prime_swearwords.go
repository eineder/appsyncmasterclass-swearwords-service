package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
)

func HandleRequest(ctx context.Context, event *any) (any, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	fmt.Printf("%+v\n", event)
	fmt.Println("Prime_Swearwords called")

	session, err := session.NewSession()
	if err != nil {
		fmt.Printf("Got error creating session: %s", err)
		return nil, err
	}

	db := dynamodb.New(session)

	errPrime := EnsureTablePrimed(db, session)
	if errPrime != nil {
		fmt.Printf("Got error priming table")
		return nil, errPrime
	}

	return nil, nil
}

func EnsureTablePrimed(db *dynamodb.DynamoDB, session *session.Session) error {
	isPrimed, err := isTablePrimed(db)
	if err != nil {
		fmt.Printf("Got error checking if table is primed")
		return err
	}

	if isPrimed {
		return nil
	}

	bucket := aws.String(os.Getenv("BUCKET_NAME"))
	key := aws.String(os.Getenv("BUCKET_KEY"))
	content, readErr := getFileContent(bucket, key, session)
	if readErr != nil {
		fmt.Printf("Got no file content")
		return readErr
	}

	// Split content into lines
	lines := strings.Split(*content, "\n")
	for _, line := range lines {
		_, err := db.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String(os.Getenv("SWEARWORDS_TABLE_NAME")),
			Item: map[string]*dynamodb.AttributeValue{
				"word": {
					S: aws.String(line),
				},
			},
		})
		if err != nil {
			fmt.Printf("Got error calling PutItem: %s", err)
			return err
		}

	}

	_, err2 := db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("SWEARWORDS_TABLE_NAME")),
		Item: map[string]*dynamodb.AttributeValue{
			"word": {
				S: aws.String("###table_primed###"),
			},
		},
	})
	if err2 != nil {
		fmt.Printf("Got error calling PutItem: %s", err2)
		return err2
	}

	return nil
}

func isTablePrimed(db *dynamodb.DynamoDB) (bool, error) {
	table_name := aws.String(os.Getenv("SWEARWORDS_TABLE_NAME"))
	result, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: table_name,
		Key: map[string]*dynamodb.AttributeValue{
			"word": {
				S: aws.String("###table_primed###"),
			},
		},
	})
	if err != nil {
		fmt.Printf("Got error calling GetItem: %s", err)
		return true, err
	}
	fmt.Println("GetItem succeeded:")
	fmt.Printf("%+v\n", result)
	return result.Item != nil, nil
}

func getFileContent(bucket *string, key *string, session *session.Session) (*string, error) {

	s3svc := s3.New(session)

	result, err := s3svc.GetObject(&s3.GetObjectInput{
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		fmt.Printf("Got error calling GetObject: %s", err)
		return nil, err
	}

	defer result.Body.Close()

	// Read the file contents into a byte slice
	fileBytes, err := ioutil.ReadAll(result.Body)
	if err != nil {
		fmt.Printf("Got error reading file: %s", err)
		return nil, err
	}

	// Convert the byte slice to a string
	fileContents := string(fileBytes)

	return &fileContents, nil
}

func main() {
	lambda.Start(HandleRequest)
}
