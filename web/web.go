package web

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/bohdantrotsenko/chainy"
)

type Server struct {
	b *chainy.Blocks
}

func New(b *chainy.Blocks) *Server {
	return &Server{b}
}

func (s *Server) serveEntry(w http.ResponseWriter, r *http.Request, e *chainy.Entry) {
	h := w.Header()
	h.Add("Content-Length", fmt.Sprintf("%d", len(e.Content)))
	if len(e.PrevHash) > 0 {
		h.Add("X-Prev-Hash", fmt.Sprintf("%x", e.PrevHash))
	}
	if len(e.NextHash) > 0 {
		h.Add("X-Next-Hash", fmt.Sprintf("%x", e.NextHash))
	}
	h.Add("X-Block", fmt.Sprintf("%d", e.Height))
	h.Add("X-Date", fmt.Sprintf("%d:%d", e.Instant.Unix(), e.Instant.Nanosecond()))
	h.Add("X-Content-Hash", fmt.Sprintf("%x", e.ContentHash))
	h.Add("X-Sign", fmt.Sprintf("%x", e.Signature))

	w.WriteHeader(http.StatusOK)
	if r.Method == "GET" {
		w.Write(e.Content)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	if r.URL == nil {
		http.Error(w, "no url", http.StatusInternalServerError)
		return
	}

	//log.Println(r.URL.Path)
	if strings.HasPrefix(r.URL.Path, "/next/") {
		hashStr := r.URL.Path[len("/next/"):]
		hash, err := hex.DecodeString(hashStr)
		if err != nil {
			http.Error(w, "bad hash", http.StatusBadRequest)
			return
		}

		entry, err := s.b.WaitForNext(r.Context(), hash)
		if err != nil {
			http.Error(w, "bad hash", http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/%x", entry.Hash()), http.StatusMovedPermanently)
		return
	}

	if r.URL.Path == "/" {
		entry, err := s.b.WaitForNext(r.Context(), nil)
		if err != nil {
			http.Error(w, "bad hash", http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/%x", entry.Hash()), http.StatusMovedPermanently)
		return
	}

	hashStr := r.URL.Path[len("/"):]
	hash, err := hex.DecodeString(hashStr)
	if err != nil {
		http.Error(w, "bad hash", http.StatusBadRequest)
		return
	}

	entry, err := s.b.Get(hash)
	if entry == nil {
		http.NotFound(w, r)
		return
	}
	s.serveEntry(w, r, entry)
}
