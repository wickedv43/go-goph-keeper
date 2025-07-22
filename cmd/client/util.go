package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"google.golang.org/grpc/metadata"
)

func (g *GophKeeper) hashPassword(password string) string {
	h := hmac.New(sha256.New, []byte(g.cfg.Master))
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

func (g *GophKeeper) authCtx() context.Context {
	if g.token == "" {
		return g.rootCtx
	}

	md := metadata.New(map[string]string{
		"authorization": "Bearer " + g.token,
	})
	return metadata.NewOutgoingContext(g.rootCtx, md)
}
