# chainy
The simplest immutable supply chain management

# Rest-like API

    GET /hash => HTTP 200 ...document
    GET /next/hash1 => HTTP 301 /hash2
    GET /next/ => HTTP 301 /epoch-hash

# Headers for /hashX

    Content-Length: <number> (mandatory)
    X-Prev-Hash: <hash> (if block > 0)
    X-Next-Hash: <hash> (if known)
    X-Block: <int64> (0..)
    X-Date: <int64>:<int64> (UTC unix epoch in seconds:nanoseconds, mandatory)
    X-Content-Hash: <hash> (mandatory)
    X-Sign: <string> (mandatory)

X-Sign must be valid for hashX

hashX is the hash of (X-Prev-Hash + X-Block + X-Date + X-Content-Hash),
where "+" means concatenation of strings,
X-Prev-Hash, X-Content-Hash are printed in hex,
X-Block is a decimal
X-Date is a decimal:decimal, just like in the header

# Hello, world

1. Create a key with `ssh-keygen -t rsa -f simple && ssh-keygen -f simple.pub -e -m pem > simple.pub.pem`.
2. Run `go run console-chain/main.go`
3. Feel free to enter entries line-by-line, they go into a 'blockchain'.
4. In a separate window, `curl -v -L http://localhost:6001/` for the first block etc.
5. Or run `go run chain-observer/main.go`