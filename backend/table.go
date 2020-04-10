package backend

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// getSess gets a session with aws
func getSess() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-2"),
		Credentials: credentials.NewSharedCredentials("", "stonesoup"),
	})
	if err != nil {
		fmt.Println("Got error calling newSession:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return sess
}

// DDBTableKey is a struct to hold the name and
// type of a table key
type DDBTableKey struct{ Name, AwsType string }

// DDBTable is a struct to hold name and keys for table
type DDBTable struct {
	Name       string
	Hkey, Rkey *DDBTableKey
}

// MakeTable will take some info and try to make the corresponding ddb table
func (table *DDBTable) MakeTable() {
	svc := dynamodb.New(getSess())
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(table.Hkey.Name),
				AttributeType: aws.String(table.Hkey.AwsType),
			},
			{
				AttributeName: aws.String(table.Rkey.Name),
				AttributeType: aws.String(table.Rkey.AwsType),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(table.Hkey.Name),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String(table.Rkey.Name),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(table.Name),
	}
	_, err := svc.CreateTable(input)
	if err != nil {
		fmt.Println("Got error calling CreateTable:")
		fmt.Println(err.Error())
		// os.Exit(1)
	} else {
		fmt.Println("Created the table", table.Name)
	}
	time.Sleep(2 * time.Second)
}

// AddItem will add a struct item to the table
func (table *DDBTable) AddItem(dst interface{}) {
	name := table.Name
	item, err := dynamodbattribute.MarshalMap(dst)
	if err != nil {
		fmt.Println("Got error marshalling new movie item:")
		fmt.Println(err.Error())
		// os.Exit(1)
	}
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(name),
	}
	ddb := dynamodb.New(getSess())
	_, err = ddb.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		// os.Exit(1)
	}
}

func getField(dst *interface{}, field string) string {
	r := reflect.ValueOf(dst)
	f := reflect.Indirect(r).FieldByName(field)
	return string(f.String())
}

// ReadItem takes a value for the hash key and value for range
// key and looks up the entry in the table
func (table *DDBTable) ReadItem(hk, rk string, dst interface{}) {
	ddb := dynamodb.New(getSess())
	result, err := ddb.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(table.Name),
		Key: map[string]*dynamodb.AttributeValue{
			table.Hkey.Name: {
				S: aws.String(hk),
			},
			table.Rkey.Name: {
				S: aws.String(rk),
			},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = dynamodbattribute.UnmarshalMap(result.Item, &dst)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}
}

// Lookup takes a key and value and return results
func (table *DDBTable) Lookup(k, v string) []map[string]*dynamodb.AttributeValue {
	ddb := dynamodb.New(getSess())
	filt := expression.Name(k).Equal(expression.Value(v))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		fmt.Println(err)
	}
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(table.Name),
	}
	// Make the DynamoDB Query API call
	result, err := ddb.Scan(input)
	if err != nil {
		fmt.Println("LookupHK call failed:")
		fmt.Println((err.Error()))
		os.Exit(1)
	}
	return result.Items
}

// Delete takes a key and value to delete and returns response
func (table *DDBTable) Delete(hk, rk string) error {
	ddb := dynamodb.New(getSess())
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			table.Hkey.Name: {
				S: aws.String(hk),
			},
			table.Rkey.Name: {
				S: aws.String(rk),
			},
		},
		TableName: aws.String(table.Name),
	}

	_, err := ddb.DeleteItem(input)
	if err != nil {
		fmt.Println("Got error calling DeleteItem")
		fmt.Println(err.Error())
	}
	return err
}

// TableBuilder is to be used by the api to make a new table struct
// it returns a pointer to the newly allocd table struct
func TableBuilder(name, hk, hkt, rk, rkt string) *DDBTable {
	hashk := &DDBTableKey{hk, hkt}
	rangek := &DDBTableKey{rk, rkt}
	newTable := &DDBTable{name, hashk, rangek}
	return newTable
}

///////////////////////
// individual tables //
///////////////////////

// MakeUserTable builds a UserTable struct and returns a pointer to it
func MakeUserTable(create bool) *DDBTable {
	table := TableBuilder("smelltest-users", "ID", "S", "ID2", "S")
	if create {
		table.MakeTable()
	}
	return table
}

// MakeReverseLookupTable builds a ReverseLookupTable struct and returns a pointer to it
func MakeReverseLookupTable(create bool) *DDBTable {
	table := TableBuilder("smelltest-reverselookup", "ReverseKey", "S", "ReverseValue", "S")
	if create {
		table.MakeTable()
	}
	return table
}

// MakeSmellEntry builds a SmellEntry struct and returns a pointer to it
func MakeSmellEntry(create bool) *DDBTable {
	table := TableBuilder("smelltest-smellentry", "ID", "S", "UID", "S")
	if create {
		table.MakeTable()
	}
	return table
}

// MainMakeAllTables should be called by main to make all tables
func MainMakeAllTables() {
	funcs := []func(bool) *DDBTable{MakeUserTable, MakeReverseLookupTable, MakeSmellEntry}
	for _, fn := range funcs {
		go fn(true)
	}
}
