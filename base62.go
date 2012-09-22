package base62

import (
  "io"
  "log"
)

type reader struct {
  r   io.Reader
  buf [1024]byte
  off uint // in bits
  num uint // in bits
  err error
}

var masks = [8]byte{0x00, 0x01, 0x03, 0x07, 0x1f, 0x3f, 0x7f, 0xff}

// Fill up internal buffers from the Reader
func (r *reader) fill() bool {
  if r.off >= r.num {
    if r.err != nil {
      return false
    }

    // TODO(knorton): Handle case where r.r is nil
    n, err := io.ReadAtLeast(r.r, r.buf[:], 1)
    r.off = 0
    r.num = uint(n) * 8
    r.err = err
    return r.num > 0
  }
  return true
}

// Reads n bits from the underlying reader.
func (r *reader) read(n uint) (uint, uint, error) {
  var rv uint = 0
  var rn uint = 0

  for n > 0 {
    if !r.fill() {
      return rv, rn, r.err
    }

    // index of buf to use
    q := r.off >> 3

    // how many bits have been read in buf[q]
    i := r.off & 7

    // bits left in buf[q]
    j := 8 - i

    log.Printf("off = %d, num = %d, q = %d, i = %d, j = %d, buf[q] = %d, buf = %v\n", r.off, r.num, q, i, j, r.buf[q], r.buf)
    // can we satisfy n from the buf[q]?
    if n <= j {
      v := uint(masks[n] & (r.buf[q] >> uint(j-n)))
      rv <<= n
      rv |= v
      rn += n
      r.off += n
      return rv, rn, r.err
    }

    v := uint(masks[j] & r.buf[q])
    rv |= v << rn
    rn += j
    r.off += j
    n -= j
  }

  panic("unreachable")
}

// type writer struct {
//   buf []byte
//   n   int
//   w   io.Writer
// }

// // Write these bits into the buffer
// func (w *writer) writeBits(v, int) {
// }

// func (w *writer) commit() (n, error) {
//   return 0, nil
// }

// const symbols = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// type Encoding struct {
//   sym string
//   pos map[byte]int
// }

// func NewEncoding(symbols string) *Encoding {
//   e := new(Encoding)
//   e.sym = symbols
//   pos := make(map[byte]int)
//   for i := 0; i < len(symbols); i++ {
//     pos[symbols[i]] = i
//   }
//   e.pos = pos
//   return e
// }

// func (e *Encoding) Encode(dst, src []byte) {
// }

// func (e *Encoding) Decode(dst, src []byte) (int, error) {
//   return 0, nil
// }

// func NewDecoder(r io.Reader) io.Reader {
//   return nil
// }

// func NewEncoder(w io.Writer) io.Writer {
//   return nil
// }

// func EncodeToString(src []byte) string {
//   return nil
// }

// func DecodeString(src string) ([]byte, error) {
//   return nil, nil
// }
