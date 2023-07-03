package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// From https://gist.github.com/6220119/7ca4244528ac65abef3a39c8a2ec7ea3

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func generateRandomStringURLSafe(n int) (string, error) {
	b, err := generateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}

func main() {
	s, err := generateRandomStringURLSafe(32)
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
}
