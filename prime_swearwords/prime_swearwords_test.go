package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type PrimeSwearwordsTestSuite struct {
	suite.Suite
}

// this function executes before the test suite begins execution
func (suite *PrimeSwearwordsTestSuite) SetupSuite() {
	fmt.Println(">>> Setup Test Suite")
	os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
}

// this function executes after all tests executed
func (suite *PrimeSwearwordsTestSuite) TearDownSuite() {
	fmt.Println(">>> Tear down Test Suite")
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(PrimeSwearwordsTestSuite))
}

func (ts *PrimeSwearwordsTestSuite) TestEnsureTablePrimed() {

	session, sessionErr := session.NewSession()
	if sessionErr != nil {
		ts.Fail("Got error creating session.", sessionErr)
	}
	db := dynamodb.New(session)
	godotenv.Load("../.env")
	tableName := os.Getenv("SWEARWORDS_TABLE_NAME")

	emptyErr := Given_AnEmptySwearwordsTable(ts, tableName, db)
	if emptyErr != nil {
		ts.Fail("Got error emptying the table.", emptyErr)
	}

	primedErr := EnsureTablePrimed(db, session)
	if primedErr != nil {
		ts.Fail("Got error priming the table:", primedErr)
	}
	response, getMarkerErr := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"word": {
				S: aws.String("###table_primed###"),
			},
		},
	})

	if getMarkerErr != nil {
		ts.Fail("Got error getting the marker item.", getMarkerErr)
	}
	if response.Item == nil {
		ts.Fail("Expected to find marker item in table")
	}

}

func Given_AnEmptySwearwordsTable(t *PrimeSwearwordsTestSuite, tableName string, db *dynamodb.DynamoDB) error {

	// Scan the table
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}
	scanOutput, err := db.Scan(scanInput)
	if err != nil {
		return err
	}

	// Prepare the batch of items to delete
	var writeRequests []*dynamodb.WriteRequest
	for _, item := range scanOutput.Items {
		writeRequests = append(writeRequests, &dynamodb.WriteRequest{
			DeleteRequest: &dynamodb.DeleteRequest{
				Key: item,
			},
		})

		// If we've collected 25 items, delete them in a batch
		if len(writeRequests) == 25 {
			_, err := db.BatchWriteItem(&dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]*dynamodb.WriteRequest{
					tableName: writeRequests,
				},
			})
			if err != nil {
				return err
			}

			// Clear the batch
			writeRequests = writeRequests[:0]
		}
	}

	// Delete any remaining items
	if len(writeRequests) > 0 {
		_, err := db.BatchWriteItem(&dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]*dynamodb.WriteRequest{
				tableName: writeRequests,
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
