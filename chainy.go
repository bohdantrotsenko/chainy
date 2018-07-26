package chainy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrEntryNotFound = errors.New("no entry with the given hash")
var ErrIncorrectTimestamp = errors.New("incorrect timestamp")
var ErrMissingSigner = errors.New("*Blocks needs a signer")
var ErrMissingVerifier = errors.New("*Blocks needs a verifier")
var ErrVerificationFailed = errors.New("crypto verification failed")

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
		panic("time must be in UTC()") // that's a bit bold, but anyway
	}
	// TODO: cache the computation
	sec := e.Instant.Unix()
	nano := e.Instant.Nanosecond()
	digest := sha256.Sum256([]byte(fmt.Sprintf("%x%d%d:%d%x%x", e.PrevHash, e.Height, sec, nano, e.Nonce, e.ContentHash)))
	return digest[:]
}

type SignerFunc func([]byte) ([]byte, error)
type VerifierFunc func(hash []byte, sign []byte) bool

// Blocks are the "chain".
type Blocks struct {
	Entries []Entry

	Signer   SignerFunc
	Verifier VerifierFunc

	modified chan struct{}
	sync.RWMutex
}

// WaitForNext returns the next entry after the given one
func (b *Blocks) WaitForNext(ctx context.Context, hash []byte) (*Entry, error) {
	b.RLock()
	m := b.modified

	if len(hash) == 0 && len(b.Entries) > 0 { // it's a request for the first block, we have it
		res := &b.Entries[0]
		b.RUnlock()
		return res, nil
	}

	if len(hash) == 0 && len(b.Entries) == 0 { // it's a request for the first block, we wait
		b.RUnlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-m: // here's the blocking (waiting) call
			return &b.Entries[0], nil
		}
	}

	var cursor *Entry
	var curIdx int
	for idx, el := range b.Entries {
		if bytes.Equal(hash, el.Hash()) { // TODO: optimize via map
			cursor = &el
			curIdx = idx
			break
		}
	}

	if cursor == nil {
		b.RUnlock()
		return nil, ErrEntryNotFound
	}

	if curIdx < len(b.Entries)-1 { // it's not the last available
		b.RUnlock()
		return &b.Entries[curIdx+1], nil
	}

	// it's the last one
	b.RUnlock()
	select {
	case <-m: // wait for the next
		return &b.Entries[curIdx+1], nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func New(signer SignerFunc, verifier VerifierFunc) *Blocks {
	return &Blocks{Signer: signer, Verifier: verifier, modified: make(chan struct{})}
}

// AppendNew creats and appends an entry to the blockchain.
func (b *Blocks) AppendNew(content []byte, instant time.Time, nonce string, signIfKnown []byte) (*Entry, error) {
	instant = instant.UTC()
	b.Lock()
	defer b.Unlock()

	if len(b.Entries) > 0 && b.last().Instant.After(instant) {
		return nil, ErrIncorrectTimestamp
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

	hash := newEntry.Hash()
	if len(signIfKnown) > 0 {
		if b.Verifier == nil {
			return nil, ErrMissingVerifier
		}
		if !b.Verifier(hash, signIfKnown) {
			return nil, ErrVerificationFailed
		}
		newEntry.Signature = signIfKnown
	} else {
		if b.Signer == nil {
			return nil, ErrMissingSigner
		}
		signature, err := b.Signer(hash)
		if err != nil {
			return nil, err
		}
		newEntry.Signature = signature
	}

	b.Entries = append(b.Entries, newEntry)
	close(b.modified)
	b.modified = make(chan struct{})

	return b.last(), nil
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
