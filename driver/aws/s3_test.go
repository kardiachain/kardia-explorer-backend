package aws

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnecttion_ConnectAws(t *testing.T) {
	session, err := ConnectAws()
	assert.Nil(t, err)
	assert.NotNil(t, session)
}
