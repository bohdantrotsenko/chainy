package main

// creates a primitive chain

import (
	"bufio"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bohdantrotsenko/chainy"
	"github.com/bohdantrotsenko/chainy/web"
)

var privateKey = flag.String("key", "simple", "private key")
var listen = flag.String("http", ":6001", "listening address")

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
	log.Println("key read into memory")

	shaSign := func(data []byte) ([]byte, error) {
		// fmt.Printf("hash to sign: %x\n", data)
		return rsa.SignPSS(rand.Reader, key, crypto.SHA256, data, nil)
	}

	bl := chainy.New(shaSign, nil)
	lis, err := net.Listen("tcp", *listen)
	if err != nil {
		return fmt.Errorf("listnener: %s", err)
	}
	defer lis.Close()

	go http.Serve(lis, web.New(bl))

	reader := bufio.NewReader(os.Stdin)

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			break
		}
		text = strings.TrimSpace(text)
		bl.AppendNew([]byte(text), time.Now().UTC(), nil)
	}

	// _, err := bl.AppendNew([]byte("first"), time.Now().UTC(), "", nil)
	// if err != nil {
	// 	return err
	// }

	j, err := json.MarshalIndent(bl.Entries, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal: %s", err)
	}

	fmt.Println(string(j))
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
