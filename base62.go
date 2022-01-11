package base62

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

var masks = []uint{0x00, 0x01, 0x03, 0x07, 0xf, 0x1f, 0x3f, 0x7f, 0xff}

var encoding = newEncoding("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")

type enc struct {
	chr string
	pos map[byte]uint
}

// TODO(knorton): Add an output buffer
type encoder struct {
	w io.Writer

	// fragment of bits left from the last write
	buf uint

	// number of bits left in buf
	num uint

	// This represents a carry state in the encoder where a single bit is
	// being pushed into the next 6-bit word. The least significant bit
	// indicates the bit value of the carry and the next significant bit
	// indicates whether there is a carry at all. Examples:
	// 0x0 - no carry
	// 0x3 - carry of 1
	// 0x2 - carry of 0
	car uint
}

type decoder struct {
	r io.Reader

	rb [1024]byte

	// fragment of written bits
	buf uint

	// bits written into buf
	num uint
}

func newEncoding(chr string) *enc {
	p := map[byte]uint{}
	for i := 0; i < len(chr); i++ {
		p[chr[i]] = uint(i)
	}
	return &enc{chr, p}
}

// Combine an bits of av with bn bits of bv to get a word with an + bn bits.
func concat(av, an, bv, bn uint) uint {
	return ((av & masks[an]) << bn) | (bv & masks[bn])
}

// Read n bits from p.
func readBits(e *encoder, n uint, p []byte) (uint, uint, []byte) {
	// log.Printf("Read %x bits\n", n)
	var rv uint = 0
	var rn uint = 0

	for n > 0 {
		// can we finish this read from buf?
		if e.num >= n {
			// read upper n bits of buf
			v := masks[n] & (e.buf >> (e.num - n))
			e.num -= n
			rv <<= n
			rv |= v
			rn += n
			return rv, rn, p
		}

		// can we make this read at all?
		if len(p) == 0 {
			return 0, 0, p
		}

		// read what is left in buf and bring in more bits from p
		v := masks[e.num] & e.buf
		rv |= v
		rn += e.num

		n -= e.num

		e.num = 8

		e.buf = uint(p[0])
		p = p[1:]
	}

	panic("unreachable")
}

// Encode a 6-bit word into a character, emit it on the writer and return
// the carry state.
func encode6Bits(e *encoder, v uint) (uint, error) {
	chr := encoding.chr
	if v < 60 {
		if _, err := e.w.Write([]byte{chr[v]}); err != nil {
			return 0, err
		}
		return 0, nil
	}
	if v < 62 {
		if _, err := e.w.Write([]byte{chr[60]}); err != nil {
			return 0, err
		}
	} else {
		if _, err := e.w.Write([]byte{chr[61]}); err != nil {
			return 0, err
		}
	}
	return (v & 1) | 2, nil
}

func (e *encoder) Write(p []byte) (int, error) {
	c := len(p)
	if c == 0 {
		return 0, nil
	}
	var v uint
	var n uint
	var err error
	for {
		if e.car&2 == 0 {
			v, n, p = readBits(e, 6, p)
			if n == 0 {
				return c, nil
			}
			e.car, err = encode6Bits(e, v)
			if err != nil {
				return c - len(p), err
			}
		} else {
			v, n, p = readBits(e, 5, p)
			if n == 0 {
				return c, nil
			}
			e.car, err = encode6Bits(e, concat(e.car&1, 1, v, 5))
			if err != nil {
				return c - len(p), err
			}
		}
	}
	panic("unreachable")
}

func (e *encoder) Close() error {
	if e.num == 0 && e.car&2 == 0 {
		return nil
	}
	if e.car&2 == 0 {
		if _, err := encode6Bits(e, e.buf&masks[e.num]<<(6-e.num)); err != nil {
			return err
		}
	} else {
		if _, err := encode6Bits(e, concat(e.car&1, 1, e.buf&masks[e.num]<<(5-e.num), 5)); err != nil {
			return err
		}
	}
	return nil
}

func writeBits(d *decoder, r *[]byte, w *[]byte, v uint, n uint) {
	// how many bits are left in our 1-byte buffer
	l := 8 - d.num

	// will we fill up buf in this write?
	if n >= l {
		(*w)[0] = byte(d.buf | (v>>(n-l))&masks[l])
		d.buf = 0
		d.num = 0

		n -= l
		l = 8
		*r = (*r)[1:]
		*w = (*w)[1:]
	}

	if n == 0 {
		return
	}

	// put what is left in buf
	d.buf |= (masks[n] & v) << (l - n)
	d.num += n
}

func (d *decoder) Read(p []byte) (int, error) {
	nn := len(p)
	nr := nn
	if nr > len(d.rb) {
		nr = len(d.rb)
	}

	q := d.rb[:nr]
	n, err := io.ReadAtLeast(d.r, q, 1)
	if err != nil && err != io.EOF {
		return 0, err
	}

	for i := 0; i < n; i++ {
		c := d.rb[i]
		pos, ok := encoding.pos[c]
		if !ok {
			return 0, errors.New(fmt.Sprintf("illegal character %d\n", c))
		}

		switch pos {
		case 60:
			writeBits(d, &q, &p, 0x1e, 5)
		case 61:
			writeBits(d, &q, &p, 0x1f, 5)
		default:
			writeBits(d, &q, &p, pos, 6)
		}
	}
	return nn - len(p), err
}

func NewEncoder(w io.Writer) io.WriteCloser {
	return &encoder{w: w}
}

func NewDecoder(r io.Reader) io.Reader {
	return &decoder{r: r}
}

func DecodedLen(n int) int {
	return n / 2
}

func EncodedLen(n int) int {
	return n * 2
}

func EncodeToString(data []byte) string {
	var buf bytes.Buffer
	e := NewEncoder(&buf)
	e.Write(data)
	e.Close()
	return buf.String()
}

func DecodeString(data string) ([]byte, error) {
	var res bytes.Buffer
	d := NewDecoder(bytes.NewBufferString(data))
	if _, err := io.Copy(&res, d); err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}
