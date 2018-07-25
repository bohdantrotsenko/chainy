package main

// reads rsa private key, prints corresponding public key

import (
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
)

var privateKey = flag.String("key", "simple", "private key")

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
	pubKey := key.Public()
	fmt.Printf("-- public key:\n%+v\n", pubKey)
	return nil
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
