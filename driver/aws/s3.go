package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"os"
)

func ConnectAws() (*session.Session, error) {
	KeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	KeyAccess := os.Getenv("AWS_SECRET_ACCESS_KEY")
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String("ap-southeast-1"),
			Credentials: credentials.NewStaticCredentials(
				KeyID,
				KeyAccess,
				"",
			),
		})

	if err != nil {
		return nil, err
	}

	return sess, nil
}
