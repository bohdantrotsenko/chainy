package main

// observes a chain

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/bohdantrotsenko/chainy"
	"github.com/bohdantrotsenko/chainy/web"
)

var publicKey = flag.String("key", "simple.pub.pem", "public key")
var target = flag.String("target", "http://localhost:6001", "target address")

func run() error {
	data, err := ioutil.ReadFile(*publicKey)
	if err != nil {
		return fmt.Errorf("reading: %s", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("no pem key found")
	}
	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse key: %s", err)
	}
	log.Println("key read into memory")

	shaVerify := func(hash []byte, sign []byte) bool {
		// fmt.Printf("hash to check: %x\n", hash)
		return rsa.VerifyPSS(key, crypto.SHA256, hash, sign, nil) == nil
	}

	bl := chainy.New(nil, shaVerify)

	go func() {
		er := web.Observe(context.Background(), *target, bl)
		if er != nil {
			log.Fatalln(er)
		}
	}()

	lastHash := make([]byte, 0)
	for {
		entry, err := bl.WaitForNext(context.Background(), lastHash)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(string(entry.Content))
		lastHash = entry.Hash()
	}

	// j, err := json.MarshalIndent(bl.Entries, "", "  ")
	// if err != nil {
	// 	return fmt.Errorf("json marshal: %s", err)
	// }

	// fmt.Println(string(j))
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
