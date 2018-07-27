package main

// creates a primitive chain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/bohdantrotsenko/chainy"
)

func shaSign(data []byte) ([]byte, error) {
	digest := sha256.Sum256(data)
	return digest[:], nil
}

func run() error {
	bl := chainy.New(shaSign, nil)

	_, err := bl.AppendNew([]byte("first"), time.Now().UTC(), nil)
	if err != nil {
		return err
	}

	j, err := json.Marshal(bl.Entries)
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
