package base62

import (
  "bytes"
  "io"
  "testing"
)

func expect(t *testing.T, v, n uint, err error) func(uint, uint, error) {
  return func(av, an uint, aerr error) {
    if v != av || n != an || err != aerr {
      t.Fatalf("Expected %d, %d, %v got %d, %d, %v", v, n, err, av, an, aerr)
    }
  }
}

func TestReader(t *testing.T) {
  // simple test
  mr := &reader{r: bytes.NewBuffer([]byte{0xff})}
  expect(t, 0x3f, 6, nil)(mr.read(6))
  expect(t, 0x3, 2, io.EOF)(mr.read(6))

  // multiple byte read
  mr = &reader{r: bytes.NewBuffer([]byte{0x6c, 0xc3})}
  expect(t, 0xd, 5, nil)(mr.read(5))
  expect(t, 0x4, 3, nil)(mr.read(3))
  expect(t, 0xc3, 8, nil)(mr.read(8))
  expect(t, 0, 0, io.EOF)(mr.read(1))

  // multiple buffer read
  d := [2048]byte{}
  for i := 0; i < 2048; i++ {
    d[i] = 0xa3
  }
  mr = &reader{r: bytes.NewBuffer(d[:])}
  for i := 0; i < 2048; i++ {
    expect(t, 0xa, 4, nil)(mr.read(4))
    expect(t, 0x3, 4, nil)(mr.read(4))
  }
  expect(t, 0, 0, io.EOF)(mr.read(4))
}
