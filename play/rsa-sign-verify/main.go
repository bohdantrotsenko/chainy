package main

// reads rsa private key, signs some data, verifies the signature

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
)

var privateKey = flag.String("key", "simple", "private key")
var input = flag.String("data", "hello, world", "something to sign")

func run() error {
	data, err := ioutil.ReadFile(*privateKey)
	if err != nil {
		return fmt.Errorf("reading: %s", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("no pem key found")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse key: %s", err)
	}

	digest := sha256.Sum256([]byte(*input))

	fmt.Printf("data: %s\n", *input)
	fmt.Printf("digest: %x\n", digest)

	signature, err := rsa.SignPSS(rand.Reader, key, crypto.SHA256, digest[:], nil)
	if err != nil {
		return fmt.Errorf("signing: %s", err)
	}
	fmt.Printf("signature: %x\n", signature)

	err = rsa.VerifyPSS(&key.PublicKey, crypto.SHA256, digest[:], signature, nil)
	if err != nil {
		return fmt.Errorf("verification failed: %s", err)
	}
	fmt.Println("verification passed.")
	return nil
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
