package main

import (
	"github.com/kardiachain/explorer-backend/api"
	"github.com/kardiachain/explorer-backend/server"
)

func main() {
	srv, err := server.New()
	if err != nil {
		panic("cannot create server instance")
	}

	api.Start(srv)
}
