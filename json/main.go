package main

import (
	"bytes"
	"errors"
	"fmt"
	"unsafe"
)

type S struct {
	b int64
	a int32
	c bool
	d string
}

func main() {
	ballast := [64]byte{}
	// s := S{}

	// fmt.Println(unsafe.Sizeof(struct {
	// 	A int64
	// 	B int32
	// 	C int32
	// }{}))

	// fmt.Println(unsafe.Sizeof(struct {
	// 	B int32
	// 	A int64
	// 	C int32
	// }{}))

	// fmt.Println(unsafe.Sizeof(s.a))
	// fmt.Println(unsafe.Sizeof(s.b))
	// fmt.Println(unsafe.Alignof(s.a))
	// fmt.Println(unsafe.Alignof(s.b))
	var b unsafe.Pointer
	b = unsafe.Pointer(&ballast)

	*(*[64]byte)(b) = [64]byte{
		3, 0, 0, 0,
		0, 0, 0, 0,
		3, 0, 0, 0,
		0, 0, 0, 0,
		1, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		1, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0}

	fmt.Println(*(*S)(b))
	parseJSON(testJSON)

	c := "aaaabbbbccccddddeeeeffffgggghhhhaaaabbbbccccddddeeeeffffgggghhhh"
	fmt.Println(unsafe.Sizeof(c))
	fmt.Println(*(*[16]byte)(unsafe.Pointer(&c)))
}

const (
	FlagReadingKey     = 0b1
	FlagReadingString  = 0b10
	FlagReadingNumber  = 0b100
	FlagReadingBoolean = 0b1000
	FlagExpectBrace    = 0b10000
)

var testJSON = []byte(`
{
	"b":1056,
	"a":123,
	"c":true
}
`)

func parseJSON(j []byte) S {
	r := bytes.NewReader(j)

	buf := make([]byte, 64)
	var err error

	detector := byte(FlagExpectBrace)

	for {
		_, err = r.Read(buf)
		var keyBuf []byte
		var valueBuf []byte
		for i := range buf {
			if i == 63 {
				break
			}

			if detector&FlagExpectBrace > 0 {
				if buf[i] != byte('{') && buf[i] != byte(' ') && buf[i] != byte('\n') && buf[i] != byte('\t') {
					err = errors.New("malformed")
					break
				}
				if buf[i] != byte('{') {
					detector ^= FlagExpectBrace
				}
				continue
			}

			if detector == 0 && buf[i] == byte('"') {
				detector ^= FlagReadingKey
				keyBuf = append(keyBuf, 0, 64)
				continue
			}

			if detector&FlagReadingKey > 0 {
				if buf[i] == byte('"') {
					detector ^= FlagReadingKey
					fmt.Println("key:", string(keyBuf))
				}
				keyBuf = append(keyBuf, buf[i])
				continue
			}

			if detector == 0 && buf[i] >= 48 && buf[i] <= 57 {
				detector ^= FlagReadingNumber
			}

			if detector^FlagReadingNumber > 0 {
				if buf[i] == byte(',') {
					detector ^= FlagReadingNumber
					fmt.Println("value:", valueBuf)
					continue
				}
				valueBuf = append(valueBuf, buf[i]-48)
				continue
			}
		}

		if err != nil {
			break
		}
		fmt.Println(string(buf))
	}
	fmt.Println(err)
	return S{}
}
