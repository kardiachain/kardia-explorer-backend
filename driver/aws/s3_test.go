package aws

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestConnecttion_ConnectAws(t *testing.T) {
	AwsAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	AwsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	AwsSecretRegion := os.Getenv("AWS_SECRET_REGION")

	session, err := ConnectAws(Config{
		KeyID:     AwsAccessKeyId,
		AccessKey: AwsSecretAccessKey,
		Region:    AwsSecretRegion,
	})
	assert.Nil(t, err)
	assert.NotNil(t, session)
}
