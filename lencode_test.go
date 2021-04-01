package lencode

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

func TestEcodeDecode(t *testing.T) {
	var buf bytes.Buffer

	byteOrders := []binary.ByteOrder{binary.BigEndian, binary.LittleEndian}

	for _, endian := range byteOrders {
		enc := NewEncoder(&buf, EndianOpt(endian))

		payload0 := make([]byte, 16)
		for i := 0; i < len(payload0); i++ {
			payload0[i] = 'a'
		}
		if err := enc.Encode(payload0); err != nil {
			t.Fatal(err)
		}

		payload1 := make([]byte, 845)
		for i := 0; i < len(payload1); i++ {
			payload1[i] = 'b'
		}
		if err := enc.Encode(payload1); err != nil {
			t.Fatal(err)
		}

		dec := NewDecoder(&buf, EndianOpt(endian))

		n, err := dec.NextLen()
		if err != nil {
			t.Fatal(err)
		}
		if n != len(payload0) {
			t.Fatalf("payload0 len mismatch %d vs %d", n, len(payload0))
		}

		got0, err := dec.Decode()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got0, payload0) {
			t.Fatalf("Payload0 mismatch %v vs %v", got0, payload0)
		}

		got1, err := dec.Decode()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got1, payload1) {
			t.Fatalf("Payload1 mismatch %v vs %v", got1, payload1)
		}

		got2, err := dec.Decode()
		if err != io.EOF {
			t.Fatalf("Expected EOF, got %s %v", err, got2)
		}
	}
}

func TestSeparatorMismatch(t *testing.T) {
	var buf bytes.Buffer

	enc := NewEncoder(&buf, SeparatorOpt(nil))

	payload0 := make([]byte, 16)
	for i := 0; i < len(payload0); i++ {
		payload0[i] = 'a'
	}
	if err := enc.Encode(payload0); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	dec := NewDecoder(r)

	_, err := dec.Decode()
	if err != separatorMismatchErr {
		t.Fatalf("Expected %s got %s", separatorMismatchErr, err)
	}

	r = bytes.NewReader(buf.Bytes())
	dec = NewDecoder(r, SeparatorOpt(nil))

	got0, err := dec.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got0, payload0) {
		t.Fatalf("Payload0 mismatch %v vs %v", got0, payload0)
	}
}

func TestDecodeUnexpectedEOF(t *testing.T) {
	payload := []byte{0x00, 0x00, 0x00, 0x01}
	r := bytes.NewReader(payload)
	dec := NewDecoder(r, SeparatorOpt(nil))

	_, err := dec.Decode()
	if err != io.ErrUnexpectedEOF {
		t.Fatalf("Short read should trigger UnexpectedEOF error but was %s", err)
	}
}
