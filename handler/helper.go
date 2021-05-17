package handler

import (
	"context"
	"fmt"

	"github.com/kardiachain/go-kaiclient/kardia"
)

func (h *handler) isKRC20(ctx context.Context, smcAddress string) {
	node := h.w.TrustedNode()
	krc20, err := kardia.NewKRC20(node, smcAddress, "")
	if err != nil {
		return
	}

	krcInfo, err := krc20.KRC20Info(ctx)
	if err != nil {
		return
	}

	fmt.Println("KRCInfo", krcInfo)
}

func (h *handler) isKRC721(ctx context.Context, smcAddress string) {
	// get KRC721 token info from RPC
	node := h.w.TrustedNode()
	krc721, err := kardia.NewKRC20(node, smcAddress, "")
	if err != nil {
		return
	}

	krcInfo, err := krc721.KRC20Info(ctx)
	if err != nil {
		return
	}

	fmt.Println("KRC721Info", krcInfo)

}
