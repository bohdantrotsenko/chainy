package web

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bohdantrotsenko/chainy"
)

// Observe monitors a url and updates the blockchain.
func Observe(ctx context.Context, url string, bl *chainy.Blocks) error {
	lastHash := bl.Last().Hash()
	cli := http.Client{}

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/next/%x", url, lastHash), nil)
		if err != nil {
			return err
		}
		resp, err := cli.Do(req.WithContext(ctx))
		if err != nil {
			return err
		}

		var content []byte
		if resp.Body != nil {
			content, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return err
			}
		}

		// nextHash := parseHex(resp.Header.Get("X-Next-Hash"), &err)
		contentHash := parseHex(resp.Header.Get("X-Content-Hash"), &err)
		date := parseDate(resp.Header.Get("X-Date"), &err)
		sign := parseHex(resp.Header.Get("X-Sign"), &err)

		if err != nil {
			return err
		}

		realContentHash := sha256.Sum256(content)
		if !bytes.Equal(contentHash, realContentHash[:]) {
			return fmt.Errorf("content doesn't match the hash")
		}

		entry, err := bl.AppendNew(content, date, sign)
		if err != nil {
			return err
		}

		lastHash = entry.Hash()
	}
}

func parseNonce(s string, err *error) string {
	if *err != nil {
		return ""
	}
	res, er := url.QueryUnescape(s)
	if er != nil {
		*err = er
		return ""
	}
	return res
}

func parseHex(s string, err *error) []byte {
	if *err != nil {
		return nil
	}
	res, er := hex.DecodeString(s)
	if er != nil {
		*err = er
		return nil
	}
	return res
}

func parseDate(s string, err *error) time.Time {
	if *err != nil {
		return time.Time{}
	}
	idx := strings.Index(s, ":")
	if idx < 0 {
		*err = fmt.Errorf("no : in sec:nsec")
		return time.Time{}
	}
	var sec, nsec int64
	sec, *err = strconv.ParseInt(s[:idx], 10, 64)
	if *err != nil {
		return time.Time{}
	}
	nsec, *err = strconv.ParseInt(s[idx+1:], 10, 64)
	if *err != nil {
		return time.Time{}
	}
	return time.Unix(sec, nsec)
}
