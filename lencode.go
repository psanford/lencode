// Package lencode implements an encoder and decoder
// for a stream of big-endian length prefixed messages.
// Messages can optionally have a static record separator
// for additional integrity checks.
package lencode // import "github.com/psanford/lencode"

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

type Encoder struct {
	w         io.Writer
	separator []byte
	lenBuf    [4]byte
	err       error
}

type Option struct {
	encoderOpt func(e *Encoder)
	decoderOpt func(e *Decoder)
}

// Specify a record separator to use as an integrity
// check when decoding. This defaults to \x6c\x65\x6e\x63.
// Set to nil to disable.
func SeparatorOpt(s []byte) Option {
	return Option{
		encoderOpt: func(e *Encoder) {
			e.separator = s
		},
		decoderOpt: func(e *Decoder) {
			e.separator = s
		},
	}
}

var defaultSeparator = []byte{'l', 'e', 'n', 'c'}
var separatorMismatchErr = errors.New("Separator mismatch")

func NewEncoder(w io.Writer, opts ...Option) *Encoder {
	e := &Encoder{
		w:         w,
		separator: defaultSeparator,
	}

	for _, opt := range opts {
		opt.encoderOpt(e)
	}
	return e
}

// Encode a message to the underlying writer. It is not safe
// to call this method concurrently.
func (e *Encoder) Encode(msg []byte) error {
	if e.err != nil {
		return e.err
	}

	if len(msg) > math.MaxUint32 {
		e.err = errors.New("Message too long to encode length in 4 bytes")
		return e.err
	}

	if e.separator != nil {
		e.write(e.separator)
	}

	binary.BigEndian.PutUint32(e.lenBuf[:], uint32(len(msg)))

	e.write(e.lenBuf[:])
	e.write(msg)

	return e.err
}

func (e *Encoder) write(b []byte) error {
	if e.err != nil {
		return e.err
	}

	_, err := e.w.Write(b)
	if err != nil {
		e.err = err
	}

	return err
}

type Decoder struct {
	r         io.Reader
	separator []byte
	prefixBuf []byte
	err       error

	pendingPrefix bool
	pendingLen    int
}

func NewDecoder(r io.Reader, opts ...Option) *Decoder {
	d := &Decoder{
		r:         r,
		separator: defaultSeparator,
	}

	for _, opt := range opts {
		opt.decoderOpt(d)
	}

	d.prefixBuf = make([]byte, len(d.separator)+4)

	return d
}

// Decode the next message from the io.Reader.
func (d *Decoder) Decode() ([]byte, error) {
	if err := d.readPrefix(); err != nil {
		return nil, err
	}

	buf := make([]byte, d.pendingLen)

	if err := d.DecodeInto(buf); err != nil {
		return nil, err
	}

	return buf, nil
}

// Decode the next message into a provided byte slice.
// Use NextLen() to ensure the slice is large enough
// for the message.
func (d *Decoder) DecodeInto(b []byte) error {
	if err := d.readPrefix(); err != nil {
		return err
	}

	if d.pendingLen < len(b) {
		d.err = errors.New("Buffer not large enough for next message")
		return d.err
	}

	if _, err := io.ReadFull(d.r, b[:d.pendingLen]); err != nil {
		d.err = err
		return d.err
	}

	d.pendingPrefix = false
	d.pendingLen = 0

	return nil
}

// Return the length of the next message
func (d *Decoder) NextLen() (int, error) {
	if d.err != nil {
		return 0, d.err
	}
	if !d.pendingPrefix {
		if err := d.readPrefix(); err != nil {
			return 0, err
		}
	}

	return d.pendingLen, nil
}

func (d *Decoder) readPrefix() error {
	if d.err != nil {
		return d.err
	}

	if d.pendingPrefix {
		return nil
	}

	if _, err := io.ReadFull(d.r, d.prefixBuf); err != nil {
		d.err = err
		return err
	}

	if len(d.separator) > 0 {
		if !bytes.Equal(d.prefixBuf[:len(d.separator)], d.separator) {
			d.err = separatorMismatchErr
			return d.err
		}
	}

	l := binary.BigEndian.Uint32(d.prefixBuf[len(d.separator):])

	d.pendingPrefix = true
	d.pendingLen = int(l)
	return d.err
}
