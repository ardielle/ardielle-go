// Copyright 2015 Yahoo Inc.
// Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.

//
// Go implementation of the tbin encoding format
//

package tbin

import (
	"bufio"
	"bytes"
	"github.com/ardielle/ardielle-go/rdl"
	"io"
)

//
// TBinMarshallable - if an object implements this interface, it is used to marshal objects to TBin.
// This is optional, a default reflection-based encoder will be used for structs that do not
// implement this.
//
type TBinMarshallable interface {
	MarshalTBin(*Encoder) error
}

//
// Marshal - Marshal the specified data to TBin, returning a byte array or error.
//
func Marshal(data interface{}) ([]byte, error) {
	enc := NewEncoder(nil)
	enc.Encode(data)
	return enc.Bytes(), enc.Error()
}

//
// Encoder - the state for the encoder.
//
type Encoder struct {
	out       io.Writer
	buf       bytes.Buffer
	reg       rdl.TypeRegistry
	err       error
	syms      map[string]int
	tags      map[string]*tagDef
	nextTag   int
	nextSymId int
	tagged    bool
	bytebuf   []byte
}

// NewEncoder - create and return a new Encoder. This is a "session" for tbin, i.e. accumulated
// state for this encoder can make repeated Marshal calls more efficient.
func NewEncoder(w io.Writer) *Encoder {
	enc := Encoder{syms: make(map[string]int, 0), tags: make(map[string]*tagDef, 0), nextTag: FirstUserTag}
	enc.out = w
	enc.bytebuf = make([]byte, 32)
	enc.writeHeader()
	return &enc
}

//
// TBinUnmarshallable - implement this interface to take control of decoding.
//
type TBinUnmarshallable interface {
	UnmarshalTBin(dec *Decoder) error
}

//
// Unmarshal - decode the TBin byte array into the specified target entity. If the
// target entity is a pointer to an interface{}, it will decode into the appropriate
// primitive types, along with map[string]interface{} for any structs and maps, and
// []interface{} for any arrays.
// If a pointer to a particular struct type is provided, it is filled with data as best
// it can, allocating substructure as needed.
// In this, it tries to imitate the encoding/json behavior.
func Unmarshal(b []byte, data interface{}) error {
	in := bytes.NewReader(b)
	decoder := NewDecoder(in)
	return decoder.Decode(data)
}

//
// Decoder - the state for the decoder
//
type Decoder struct {
	dataVersion int
	//types         []*tagType
	types        []*Signature
	syms         []string
	currentTag   int
	currentCount int
	//	currentCursor *structCursor
	err        error
	pendingTag int
	in         *bufio.Reader
}

// NewDecoder - create and return a new Encoder. This is a "session" for tbin, i.e. accumulated
// state for this encoder can make repeated Marshal calls more efficient.
func NewDecoder(r io.Reader) *Decoder {
	decoder := new(Decoder)
	decoder.pendingTag = -1
	decoder.syms = make([]string, 0)
	decoder.in = bufio.NewReader(r)
	//	decoder.currentCursor = nil
	decoder.readHeader()
	return decoder
}
