// Copyright (c) 2014, Kevin Walsh.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

type NopWriteCloser struct {
	c io.Writer
}

func (w *NopWriteCloser) Write(p []byte) (n int, err error) {
	return w.c.Write(p)
}

func (w *NopWriteCloser) Close() error {
	return nil
}

func PairHelper(t *testing.T, closeA, closeB bool, msg string) {
	a, _ := ioutil.TempFile(os.TempDir(), "tempA")
	defer os.Remove(a.Name())
	b, _ := ioutil.TempFile(os.TempDir(), "tempB")
	defer os.Remove(b.Name())

	var rA io.ReadCloser
	if closeA {
		rA = a
	} else {
		rA = ioutil.NopCloser(a)
	}
	var rB io.WriteCloser
	if closeB {
		rB = b
	} else {
		// rB = ioutil.NopCloser(b) // argh: doesn't exist?
		rB = &NopWriteCloser{b}
	}

	p := NewPairReadWriteCloser(rA, rB)
	_, err := io.WriteString(p, "hello")
	if err != nil {
		t.Error(err.Error)
		return
	}
	err = p.Close()
	if err != nil {
		t.Error(err.Error)
		return
	}
	if p.Close() == nil && (closeA || closeB) {
		t.Error("pair did not detect double close for " + msg)
	}
	if (a.Close() != nil) != (closeA) {
		t.Error("tempA close was not handled properly for " + msg)
	}
	if (b.Close() != nil) != (closeB) {
		t.Error("tempB close was not handled properly for " + msg)
	}

	a, _ = os.Open(a.Name())
	b, _ = os.Open(b.Name())

	p = NewPairReadWriteCloser(b, a)
	buf := make([]byte, 100)
	n, err := io.ReadFull(p, buf)
	if err != io.ErrUnexpectedEOF {
		t.Error(err.Error())
		return
	}
	if n != len("hello") {
		t.Error("wrong number of chars")
	}
	p.Close()
}

func TestPairReadWriteCloser(t *testing.T) {
	PairHelper(t, true, true, "double close")
	PairHelper(t, true, false, "close only A")
	PairHelper(t, false, true, "close only B")
	PairHelper(t, false, false, "close neither")
}
