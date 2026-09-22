// Sshwifty - A Web SSH client
//
// Copyright (C) 2019-2025 Ni Rui <ranqus@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package commands

import (
	"errors"

	"github.com/nirui/sshwifty/application/rw"
)

// Errors
var (
	ErrStringParseBufferTooSmall = errors.New(
		"not enough buffer space to parse given string")

	ErrStringMarshalBufferTooSmall = errors.New(
		"not enough buffer space to marshal given string")
)

// String data
type String struct {
	len  Integer
	data []byte
}

// ParseString build the String according to currently read data
func ParseString(reader rw.ReaderFunc, b []byte) (String, Integer, error) {
	lenData := Integer(0)
	mErr := lenData.Unmarshal(reader)
	if mErr != nil {
		return String{}, 0, mErr
	}
	bLen := len(b)
	if bLen < lenData.Int() {
		return String{}, 0, ErrStringParseBufferTooSmall
	}
	_, rErr := rw.ReadFull(reader, b[:lenData])
	if rErr != nil {
		return String{}, 0, rErr
	}
	return String{
		len:  lenData,
		data: b[:lenData],
	}, lenData, nil
}

// Errors for NewString
var (
	errStringDataTooLong = errors.New(
		"Data was too long for a String")
)

// NewString create a new String
func NewString(d []byte) String {
	dLen := len(d)
	if dLen > MaxInteger {
		panic(errStringDataTooLong.Error())
	}
	return String{
		len:  Integer(dLen),
		data: d,
	}
}

// BuildString creates a string based on given `d`
func BuildString(d []byte) (String, error) {
	dLen := len(d)
	if dLen > MaxInteger {
		return String{}, errStringDataTooLong
	}
	return String{
		len:  Integer(dLen),
		data: d,
	}, nil
}

// Data returns the data of the string
func (s String) Data() []byte {
	return s.data
}

// Marshal the string to give buffer
func (s String) Marshal(b []byte) (int, error) {
	bLen := len(b)
	if bLen < s.len.ByteSize()+len(s.data) {
		return 0, ErrStringMarshalBufferTooSmall
	}
	mLen, mErr := s.len.Marshal(b)
	if mErr != nil {
		return 0, mErr
	}
	return copy(b[mLen:], s.data) + mLen, nil
}

// MarshalString marshals `s` into `b`
func MarshalString(s string, b []byte) (int, error) {
	if s, err := BuildString([]byte(s)); err != nil {
		return 0, err
	} else {
		return s.Marshal(b)
	}
}

// Errors for MarshalStrings
var (
	errMarshalStringsTooManyStrings = errors.New(
		"too many strings to marshal")
)

// MarshalString marshals `s` into `b`
func MarshalStrings(s []string, b []byte) (int, error) {
	if len(s) > MaxInteger {
		return 0, errMarshalStringsTooManyStrings
	}
	size := Integer(len(s))
	start, err := size.Marshal(b)
	if err != nil {
		return 0, err
	}
	for i := range s {
		if n, err := MarshalString(s[i], b[start:]); err != nil {
			return 0, err
		} else {
			start += n
		}
	}
	return start, nil
}

// ParseStrings parses strings read from given read
func ParseStrings(reader rw.ReaderFunc, b []byte) ([]String, Integer, error) {
	size := Integer(0)
	if err := size.Unmarshal(reader); err != nil {
		return nil, 0, err
	}
	out := make([]String, 0, size)
	next := Integer(0)
	for range size {
		if s, consumed, err := ParseString(reader, b[next:]); err != nil {
			return nil, 0, err
		} else {
			next += consumed
			out = append(out, s)
		}
	}
	return out, next, nil
}
