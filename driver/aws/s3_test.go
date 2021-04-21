package aws

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnecttion_ConnectAws(t *testing.T) {
	s3 := &S3{}
	session, err := s3.ConnectAws()
	assert.Nil(t, err)
	assert.NotNil(t, session)
}
