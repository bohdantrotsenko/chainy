# chainy
The simplest immutable supply chain management

# Rest-like API

    GET /hash => HTTP 200 ...document
    GET /next/hash1 => HTTP 301 /hash2
    GET /prev/hash1 = HTTP 301 /hash2
    GET /next/ => HTTP 301 /epoch-hash

# Headers for /hashX

    Content-Length: <number> (mandatory)
    X-Prev-Hash: <hash> (if block > 0)
    X-Next-Hash: <hash> (if known)
    X-Block: <int64> (0..)
    X-Date: <int64>:<int64> (UTC unix epoch in seconds:nanoseconds, mandatory)
    X-Nonce: <string 0-160 chars> (mandatory)
    X-Content-Hash: <hash> (mandatory)
    X-Sign: <string> (mandatory)

hashX is the hash of (X-Prev-Hash + X-Block + X-Date + X-Nonce + X-Content-Hash)
X-Sign must be valid for hashX

# Helpful commands

create key:
ssh-keygen -t rsa -f simple
