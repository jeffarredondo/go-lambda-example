package api

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

//AssumeRoleWSession assumes a role in AWS
func AssumeRoleWSession(accountID, accountName, roleName, sessionName string, sess *session.Session) (*session.Session, error) {
	params := &sts.AssumeRoleInput{
		RoleArn:         buildRoleArn(accountID, roleName),
		RoleSessionName: buildSessionName(accountName, sessionName),
	}
	svc := sts.New(sess)
	resp, err := svc.AssumeRole(params)
	if err != nil {
		return nil, err
	}
	return session.New(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			*resp.Credentials.AccessKeyId,
			*resp.Credentials.SecretAccessKey,
			*resp.Credentials.SessionToken),
	}), nil
}

// AssumeRoleWMFA use default credentials to assume role in another account
func AssumeRoleWMFA(accountID, accountName, roleName, sessionName, tokenCode, serialNumber string) (*session.Session, error) {
	svc := sts.New(session.New())
	resp, err := svc.AssumeRole(&sts.AssumeRoleInput{
		RoleArn:         buildRoleArn(accountID, roleName),
		RoleSessionName: buildSessionName(accountName, sessionName),
		SerialNumber:    aws.String(serialNumber),
		TokenCode:       aws.String(tokenCode),
	})
	if err != nil {
		return nil, err
	}
	return session.New(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			*resp.Credentials.AccessKeyId,
			*resp.Credentials.SecretAccessKey,
			*resp.Credentials.SessionToken),
	}), nil
}

//buildRoleArn name will build the role arn
func buildRoleArn(accountID, roleName string) *string {
	s := fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, roleName)
	return &s
}

//buildSessionName will build the session name for logging purposes
func buildSessionName(alias, sessionName string) *string {
	s := fmt.Sprintf("%vBuildVirtualInterfaces", alias)
	return &s
}

//DescribeInstances lists all instances in a given region
func DescribeInstances(sess *session.Session, region string) ([]string, []string, error) {

	var instances []string
	var keyNames []string

	//pass in session created from assume role to create new ec2 client/service
	svc := ec2.New(sess, aws.NewConfig().WithRegion(region))

	//create api request input, since we want everything we'll leave it blank
	// but this can filter to specific instance IDs
	// https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#DescribeInstancesInput
	input := &ec2.DescribeInstancesInput{}

	//hit aws api with request
	// https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#DescribeInstancesOutput
	results, err := svc.DescribeInstances(input)
	if err != nil {
		log.Println(err)
		return instances, keyNames, errors.New("error describing ec2 instances")
	}

	// iterate through results. The result contains a token to paginate (skipping but when in doubt paginate)
	// and also an array of type Reservation. https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#Reservation
	for _, reservation := range results.Reservations {

		// in that array is an array of instances containing one or more instances types
		//	https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#Instance
		for _, instance := range reservation.Instances {
			instances = append(instances, *instance.InstanceId)
			keyNames = append(keyNames, *instance.KeyName)
		}
	}

	return instances, keyNames, nil

}
