package chainy

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Entry in an entry in the chain.
type Entry struct {
	Instant     time.Time
	Height      int
	PrevHash    []byte
	Nonce       string
	ContentHash []byte
	Content     []byte
	Signature   []byte
}

// Hash computes the reference hash for the entry.
func (e *Entry) Hash() []byte {
	if e.Instant.Location() != time.UTC {
		panic("time must be in UTC()")
	}
	sec := e.Instant.Unix()
	nano := e.Instant.Nanosecond()
	digest := sha256.Sum256([]byte(fmt.Sprintf("%x%d%d:%d%x%x", e.PrevHash, e.Height, sec, nano, e.Nonce, e.ContentHash)))
	return digest[:]
}

// Blocks are the "chain".
type Blocks struct {
	Entries []Entry

	Signer   func([]byte) ([]byte, error)
	Verifier func([]byte) bool

	sync.RWMutex
}

func (b *Blocks) last() *Entry {
	return &b.Entries[len(b.Entries)-1]
}

func (b *Blocks) lastHash() []byte {
	if len(b.Entries) == 0 {
		return make([]byte, 0)
	}
	return b.last().Hash()
}

func clone(data []byte) []byte {
	res := make([]byte, len(data))
	copy(res, data)
	return res
}

var ErrIncorrectTimestamp = errors.New("incorrect timestamp")

func (b *Blocks) AppendNew(content []byte, instant time.Time, nonce string) error {
	instant = instant.UTC()
	b.Lock()
	defer b.Unlock()

	if len(b.Entries) > 0 && b.last().Instant.After(instant) {
		return ErrIncorrectTimestamp
	}

	contentDigest := sha256.Sum256(content)
	newEntry := Entry{
		Instant:     instant,
		Height:      len(b.Entries),
		PrevHash:    b.lastHash(),
		Nonce:       nonce,
		ContentHash: contentDigest[:],
		Content:     clone(content),
	}

	signature, err := b.Signer(newEntry.Hash())
	if err != nil {
		return err
	}

	newEntry.Signature = signature
	b.Entries = append(b.Entries, newEntry)

	return nil
}
