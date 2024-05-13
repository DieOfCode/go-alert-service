package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func sign(v any, key string) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	h1 := hmac.New(sha256.New, []byte(key))
	h1.Write(b)
	d := h1.Sum(nil)
	return hex.EncodeToString(d)
}
