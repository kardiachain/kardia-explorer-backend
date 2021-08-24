// Package cache
package cache

import (
	"context"
)

const (
	KeyTokenPreprocess = "token#preprocess"
)

type IToken interface {
	PushToken(ctx context.Context, token []string) error
	PopToken(ctx context.Context) (string, error)
}

func (c *Redis) PushToken(ctx context.Context, hashes []string) error {
	var insertList []interface{}
	for _, h := range hashes {
		insertList = append(insertList, h)
	}
	_, err := c.client.LPush(ctx, KeyTokenPreprocess, insertList...).Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *Redis) PopToken(ctx context.Context) (string, error) {
	hash, err := c.client.LPop(ctx, KeyTokenPreprocess).Result()
	if err != nil {
		return "", err
	}

	return hash, nil
}
