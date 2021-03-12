// Package server
package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSrv_ImportBlock(t *testing.T) {

	srv, err := createDevSrv()
	assert.Nil(t, err)

	b, err := srv.kaiClient.BlockByHeight(context.Background(), 522934)
	assert.Nil(t, err)

	assert.Nil(t, srv.ImportBlock(context.Background(), b, true))
}
