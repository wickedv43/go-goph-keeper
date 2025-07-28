package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"google.golang.org/grpc/metadata"
)

// hashPassword returns an HMAC-SHA256 hash of the given password using the master key.
func (g *GophKeeper) hashPassword(password string) string {
	h := hmac.New(sha256.New, []byte(g.cfg.Master))
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

// authCtx returns a gRPC context with the current user's authorization token, if available.
func (g *GophKeeper) authCtx() context.Context {
	token, err := g.storage.GetCurrentToken()
	if err != nil {
		return g.rootCtx
	}

	if token == "" {
		return g.rootCtx
	}

	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	return metadata.NewOutgoingContext(g.rootCtx, md)
}

// printBanner prints the ASCII banner and build information to the console.
func (g *GophKeeper) printBanner() {
	fmt.Print(`
   _____             _     _  __                         
  / ____|           | |   | |/ /                         
 | |  __  ___  _ __ | |__ | ' / ___  ___ _ __   ___ _ __ 
 | | |_ |/ _ \| '_ \| '_ \|  < / _ \/ _ | '_ \ / _ | '__|
 | |__| | (_) | |_) | | | | . |  __|  __| |_) |  __| |   
  \_____|\___/| .__/|_| |_|_|\_\___|\___|\___| .__/ \___| 
              | |                       | |              
              |_|                       |_|              
`)
	fmt.Printf("📦  Версия: %s | 📅  Сборка: %s\n\n",
		buildVersion, buildDate)
}
