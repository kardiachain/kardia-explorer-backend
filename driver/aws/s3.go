package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type ConfigAccessAWS struct {
	KeyID     string
	KeyAccess string
	Region    *string
}

func ConnectAws(config ConfigAccessAWS) (*session.Session, error) {
	KeyID := config.KeyID
	KeyAccess := config.KeyAccess
	sess, err := session.NewSession(
		&aws.Config{
			Region: config.Region,
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
