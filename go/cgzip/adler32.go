// +build cgo

/*
Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cgzip

/*
#cgo CFLAGS: -Werror=implicit
#cgo LDFLAGS: -lz
#cgo pkg-config: zlib

#include "zlib.h"
*/
import "C"

import (
	"hash"
	"unsafe"
)

type adler32Hash struct {
	adler C.uLong
}

// NewAdler32 creates an empty buffer which has an adler32 of '1'. The go
// hash/adler32 does the same.
func NewAdler32() hash.Hash32 {
	a := &adler32Hash{}
	a.Reset()
	return a
}

// io.Writer interface
func (a *adler32Hash) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		a.adler = C.adler32(a.adler, (*C.Bytef)(unsafe.Pointer(&p[0])), (C.uInt)(len(p)))
	}
	return len(p), nil
}

// hash.Hash interface
func (a *adler32Hash) Sum(b []byte) []byte {
	s := a.Sum32()
	b = append(b, byte(s>>24))
	b = append(b, byte(s>>16))
	b = append(b, byte(s>>8))
	b = append(b, byte(s))
	return b
}

func (a *adler32Hash) Reset() {
	a.adler = C.adler32(0, (*C.Bytef)(unsafe.Pointer(nil)), 0)
}

func (a *adler32Hash) Size() int {
	return 4
}

func (a *adler32Hash) BlockSize() int {
	return 1
}

// hash.Hash32 interface
func (a *adler32Hash) Sum32() uint32 {
	return uint32(a.adler)
}
