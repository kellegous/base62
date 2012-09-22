package base62

import (
  "bytes"
  "io"
  "log"
  "testing"
)

func TestReader(t *testing.T) {
  mr := &reader{r: bytes.NewBuffer([]byte{0xff})}

  v, n, err := mr.read(6)
  if v != 0x3f || n != 6 || err != nil {
    t.Errorf("Expected 0x3f, 6, nil; got %d, %d, %v", v, n, err)
  }

  log.Printf("%d, %d, %v\n", v, n, err)
  v, n, err = mr.read(6)
  if v != 0x3 || n != 2 || err != io.EOF {
    t.Errorf("Expected 0x2, 2, EOF; got %d, %d, %v", v, n, err)
  }
  log.Printf("%d, %d, %v\n", v, n, err)
}
