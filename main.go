package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jeffarredondo/go-lambda-example/api"
)

//Handler is used by the lambda function
func Handler() {
	accountId := os.Getenv("AccountId")
	accountName := os.Getenv("AccountName")
	roleName := os.Getenv("RoleName")
	region := os.Getenv("Region")

	//starting default session in account lambda is in
	defaultSession := session.New()

	//creating new session in host account by assuming a given role. If instances are in same account, delete this
	// and change the next reference to hostSession to default session.
	hostSession, err := api.AssumeRoleWSession(accountId, accountName, roleName, "DescribeInstancesSession", defaultSession)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// use api wrappe you created
	instances, keynames, err := api.DescribeInstances(hostSession, region)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	//dump lists to logs
	log.Println(instances)
	log.Println(keynames)

}

func main() {
	//need to start handler for lambda
	lambda.Start(Handler)
}
