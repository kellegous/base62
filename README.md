# Base62 Encoder/Decoder
## a [Go](http://golang.org/) library to encode and decode byte streams to a base62 encoding.

Roughly based on parts of [A Secure, Lossless, and Compressed Base62
Encoding](http://www.molengo.com/dl/base62_encoding.pdf), the encoding scheme does
not use the traditional approach requiring a lot of modulo math and is therefore
considerably more efficient with minimal loss in compression (~0.28%).

Base62 is an encoding that is mostly associated with URL shorteners. It is much
like the widely used Base64 encoding, but constrains the character set to just
the 62 alphanumeric characters.

## Installation & Use
```
go get github.com/kellegous/base62
```

The encoder and decoder have APIs very similar to the one used by most Go encoding libraries (including [base64](http://golang.org/pkg/encoding/base64/)).