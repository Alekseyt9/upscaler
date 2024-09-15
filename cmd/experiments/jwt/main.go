package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func generateSecretKey() string {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(key)
}

func main() {
	secretKey := generateSecretKey()
	fmt.Println("Your secret key:", secretKey)
}
