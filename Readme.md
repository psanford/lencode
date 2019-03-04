lencode: a go length prefix encoder/decode
==========================================

The lencode package provides an encoder and decoder for reading and writing length prefixed messages to io.Reader and Writers.

Each message is prefixed with an optional separator for an additional integretry check followed by a 4 byte bigendian uint32 length.


## Documentation

https://godoc.org/github.com/psanford/lencode


## Installation

```
go get github.com/psanford/lencode
```
