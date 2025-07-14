package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func (g *GophKeeper) hashPassword(password string) string {
	h := hmac.New(sha256.New, []byte(g.cfg.Master))
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}
