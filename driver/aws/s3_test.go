package aws

import (
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnecttion_ConnectAws(t *testing.T) {
	session, err := ConnectAws()
	assert.Nil(t, err)
	assert.NotNil(t, session)

	if session != nil {
		s3 := &S3{Session: session}
		pathFile, err := s3.UploadLogo(cfg.DefaultKRCTokenLogo, "0x98BB872842FE158Dad8E8cCB8646EBDbC8AaE3B3")
		assert.Nil(t, err)
		assert.NotNil(t, pathFile)
		assert.NotEmpty(t, pathFile)
	}
}
