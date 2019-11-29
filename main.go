package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ssm"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	session2 "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ministryofjustice/analytical-platform-auth0-user-exports/auth"
	"github.com/ministryofjustice/analytical-platform-auth0-user-exports/connection"
	"github.com/ministryofjustice/analytical-platform-auth0-user-exports/env"
	"github.com/ministryofjustice/analytical-platform-auth0-user-exports/job"
	s "github.com/ministryofjustice/analytical-platform-auth0-user-exports/ssm"
)

var (
	apiUrl         = env.GetWithDefault("API_URL", "https://alpha-analytics-moj.eu.auth0.com")
	connectionName = env.GetWithDefault("CONNECTION_NAME", "github")
	platform       = os.Getenv("ENV")
	bucket         = env.GetWithDefault("BUCKET", "airflow-auth0-user-exports")
	session        = session2.Must(session2.NewSession())
	creds          = stscreds.NewCredentials(session, env.GetWithDefault("ROLE_ARN", "arn:aws:iam::593291632749:role/airflow_auth0_user_exports"))
	awsConfig      = &aws.Config{Credentials: creds, Region: aws.String("eu-west-1"), CredentialsChainVerboseErrors: aws.Bool(true)}
	s3Svc          = s3.New(session, awsConfig)
	ssmSvc         = ssm.New(session, awsConfig)
)

func main() {

	clientId := env.GetWithDefault("CLIENT_ID", "")
	clientSecret := env.GetWithDefault("CLIENT_SECRET", "")

	// if CLIENT_ID and CLIENT_SECRET env variables not set, read these from AWS parameter store
	if len(clientId) == 0 && len(clientSecret) == 0 {

		ssmPath := env.GetOrFail("SSM_PATH")
		params, err := s.GetSsmParams(ssmSvc, ssmPath)
		if err != nil {
			log.Fatalf("Error retrieving values from SSM Parameter Store: \n%s", err)
		}

		clientId = params["CLIENT_ID"]
		clientSecret = params["CLIENT_SECRET"]
	}

	tkn, err := auth.GetToken(apiUrl, clientId, clientSecret)
	if err != nil {
		log.Fatalf("Error while retrieving access token: %s", err)
	}

	connectionId, err := connection.GetConnection(fmt.Sprintf("%s/api/v2/connections", apiUrl), tkn.AccessToken, connectionName)
	if err != nil {
		log.Fatalf("Error while retrieving connection id: %s", err)
	}

	jobConfig := fmt.Sprintf(`{
		"connection_id": "%s",
		"format": "csv", 
		"fields": [
			{"name": "email"}, { "name": "nickname", "export_as": "username"}, {"name": "created_at", "export_as": "joined"}
		]
	}`, *connectionId)

	bulkUserExportJob, err := job.CreateJob(fmt.Sprintf("%s/api/v2/jobs/users-exports", apiUrl), tkn.AccessToken, jobConfig)
	if err != nil {
		log.Fatalf("Error occurred while creating job: \n%s", err)
	}

	resultLocation, err := job.WaitForJobCompletion(fmt.Sprintf("%s/api/v2/jobs", apiUrl), tkn.AccessToken, *bulkUserExportJob)
	if err != nil {
		log.Fatal(err)
	}

	exportData, err := job.GetUserExport(*resultLocation)
	if err != nil {
		log.Fatalf("Error trying to download userdata: \n%s", err)
	}

	key := fmt.Sprintf("userdata-%s.csv", time.Now().Format("02-01-2006"))

	if platform == "aws" {

		result, err := job.UploadUserExportToS3(s3Svc, exportData, bucket, key)
		if err != nil {
			log.Print(fmt.Errorf("Error while uploading userdata to S3: \n%s", err))
		}

		log.Printf("FILE=%s successfully uploaded to BUCKET=%s, VERSIONID=%p", key, bucket, result.VersionId)

	} else {

		ok, err := job.WriteLocalFile(exportData, env.GetWithDefault("FILE_PATH", "/tmp/userdata.csv"))
		if err != nil {
			log.Fatal(err)
		}

		log.Print(*ok)
	}
}
