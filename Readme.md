lencode: a go length prefix encoder/decode
==========================================

The lencode package provides an encoder and decoder for reading and writing length prefixed messages to io.Reader and Writers.

Each message is prefixed with an optional separator for an additional integretry check followed by a 4 byte (uint32) length field (defaults to BigEndian).


## Documentation

https://godoc.org/github.com/psanford/lencode

## Usage example

```

package main

import (
	"bytes"
	"encoding/hex"
	"log"
	"strings"

	"github.com/psanford/lencode"
)

func main() {
	var buf bytes.Buffer

	// create a new encoder that writes a 4 byte length (big endian)
	// followed by a message of that length
	enc := lencode.NewEncoder(&buf, lencode.SeparatorOpt(nil))

	enc.Encode([]byte("hello world\n"))
	enc.Encode([]byte("this is a second lencode message"))

	dump := hex.Dump(buf.Bytes())
	dump = strings.TrimSpace(dump)
	expect := `
00000000  00 00 00 0c 68 65 6c 6c  6f 20 77 6f 72 6c 64 0a  |....hello world.|
00000010  00 00 00 20 74 68 69 73  20 69 73 20 61 20 73 65  |... this is a se|
00000020  63 6f 6e 64 20 6c 65 6e  63 6f 64 65 20 6d 65 73  |cond lencode mes|
00000030  73 61 67 65                                       |sage|
`
	expect = strings.TrimSpace(expect)
	if dump != expect {
		log.Fatalf("got:\n%s\nexpected:\n%s\n", dump, expect)
	}

	buf.Reset()

	// create an encoder that includes a record seperator at the start of each record
	// this is useful to detect broken streams and possibly to resynchronize streams
	enc = lencode.NewEncoder(&buf, lencode.SeparatorOpt([]byte("--record_seperator--")))

	enc.Encode([]byte("hello world\n"))
	enc.Encode([]byte("this is a second lencode message"))

	dump = hex.Dump(buf.Bytes())
	dump = strings.TrimSpace(dump)
	expect = `
00000000  2d 2d 72 65 63 6f 72 64  5f 73 65 70 65 72 61 74  |--record_seperat|
00000010  6f 72 2d 2d 00 00 00 0c  68 65 6c 6c 6f 20 77 6f  |or--....hello wo|
00000020  72 6c 64 0a 2d 2d 72 65  63 6f 72 64 5f 73 65 70  |rld.--record_sep|
00000030  65 72 61 74 6f 72 2d 2d  00 00 00 20 74 68 69 73  |erator--... this|
00000040  20 69 73 20 61 20 73 65  63 6f 6e 64 20 6c 65 6e  | is a second len|
00000050  63 6f 64 65 20 6d 65 73  73 61 67 65              |code message|
`
	expect = strings.TrimSpace(expect)
	if dump != expect {
		log.Fatalf("got:\n%s\nexpected:\n%s\n", dump, expect)
	}

	buf2 := bytes.NewBuffer([]byte{
		0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o',
		0, 0, 0, 5, 'w', 'o', 'r', 'l', 'd',
	})

	dec := lencode.NewDecoder(buf2, lencode.SeparatorOpt(nil))

	msg, _ := dec.Decode()
	fmt.Printf("%s\n", msg)
	// Output: hello

	msg, _ = dec.Decode()
	fmt.Printf("%s\n", msg)
	// Output: world

	_, err := dec.Decode()
	if err != io.EOF {
		panic(err)
	}
}

```

## Installation

```
go get github.com/psanford/lencode
```
