package chainy_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/bohdantrotsenko/chainy"
	"github.com/stretchr/testify/assert"
)

func shaSign(data []byte) ([]byte, error) {
	digest := sha256.Sum256(data)
	return digest[:], nil
}

func shaVerify(hash []byte, sign []byte) bool {
	d, _ := shaSign(hash)
	return bytes.Equal(d, sign)
}

func TestChain(t *testing.T) {
	bl := chainy.New(shaSign, nil)
	inst := time.Unix(1500123123, 0)
	entry, err := bl.AppendNew([]byte("test"), inst, nil)
	assert.Nil(t, err, "no error appending an entry")
	assert.NotNil(t, entry)
}

func TestReplay(t *testing.T) {
	bl1 := chainy.New(shaSign, nil)
	inst := time.Unix(1500123123, 0)
	_, err := bl1.AppendNew([]byte("test"), inst, nil)
	assert.Nil(t, err, "no error appending an entry")

	entry := bl1.Entries[0]

	bl2 := chainy.New(nil, shaVerify)
	eTest, err := bl2.AppendNew([]byte("test"), inst, entry.Signature)
	assert.Nil(t, err, "no error appending a known entry")
	assert.NotNil(t, eTest)
}

func TestWaitFirst(t *testing.T) {
	bl := chainy.New(shaSign, nil)
	inst := time.Unix(1500123123, 0)

	errChan := make(chan error, 1)
	go func() {
		time.Sleep(3 * time.Millisecond)
		_, er := bl.AppendNew([]byte("test"), inst, nil)
		errChan <- er
	}()

	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second)
	defer ctxCancel()

	entry, err := bl.WaitForNext(ctx, nil)
	assert.Nil(t, err, "no error receiveing the entry")
	assert.Nil(t, <-errChan, "no error appending an entry")
	assert.NotNil(t, entry, "non-nil entry")
	assert.Equal(t, "test", string(entry.Content))
}

func TestWaitFirstNTimes(t *testing.T) {
	bl := chainy.New(shaSign, nil)
	inst := time.Unix(1500123123, 0)

	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second)
	defer ctxCancel()

	N := 100
	errChan := make(chan error, N)
	for i := 0; i < N; i++ {
		go func() {
			entry, err := bl.WaitForNext(ctx, nil)
			if err != nil {
				errChan <- fmt.Errorf("non-nil error in WaitForNext")
				return
			}
			if entry == nil {
				errChan <- fmt.Errorf("nil entry")
				return
			}
			if string(entry.Content) != "test" {
				errChan <- fmt.Errorf("unexpected content")
				return
			}
			errChan <- nil
		}()
	}

	time.Sleep(3 * time.Millisecond)
	_, err := bl.AppendNew([]byte("test"), inst, nil)
	assert.Nil(t, err, "no error appending an entry")

	for i := 0; i < N; i++ {
		assert.Nil(t, <-errChan)
	}
}

func TestWaitSecond(t *testing.T) {
	bl := chainy.New(shaSign, nil)

	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second)
	defer ctxCancel()

	first, err := bl.AppendNew([]byte("first"), time.Unix(1500123000, 0), nil)
	assert.Nil(t, err, "no error creating first entry")

	errChan := make(chan error, 1)
	go func() {
		time.Sleep(3 * time.Millisecond)
		_, er := bl.AppendNew([]byte("second"), time.Unix(1500123000, 0), nil)
		errChan <- er
	}()

	entry, err := bl.WaitForNext(ctx, first.Hash())
	assert.Nil(t, err, "no error receiveing the entry")
	assert.Nil(t, <-errChan, "no error appending an entry")
	assert.NotNil(t, entry, "non-nil entry")
	assert.Equal(t, "second", string(entry.Content))

	firstHash := first.Hash()
	firstHash[0] = firstHash[0] + 1 // modify it

	entry, err = bl.WaitForNext(ctx, firstHash)
	assert.Nil(t, entry, "should not be found")
	assert.NotNil(t, err, "should have an error")
}
