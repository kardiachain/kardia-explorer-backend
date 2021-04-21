package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnecttion_ConnectAws(t *testing.T) {
	session, err := ConnectAws(Config{
		KeyID:     "AKIAJI3Y5XWKQTDRL5HQ",
		KeyAccess: "GWGuKvvVnUAQCGAmY937QcKkX//0RR2SPrdh+F3w",
		Region:    aws.String("ap-southeast-1"),
	})
	assert.Nil(t, err)
	assert.NotNil(t, session)
}
