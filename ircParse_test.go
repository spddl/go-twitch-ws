// +build windows linux js,wasm

package twitch

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

func readFile(file string) []byte {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return data
}

// go test -bench=. -benchtime=20s -bench ^(BenchmarkParseIRCMessage.*)$

func BenchmarkParseIRCMessageString(b *testing.B) { // stolen from https://github.com/gempir/go-twitch-irc/blob/master/irc.go
	data := readFile("chatlog_test.log")
	for n := 0; n < b.N; n++ {
		msg := strings.Split(string(data), "\r\n")
		for _, v := range msg {
			if v == "" {
				continue
			}
			_, err := parseIRCMessageString(v)
			if err != nil {
				panic(err)
			}
		}
	}
}

func BenchmarkParseIRCMessageByte(b *testing.B) { // convert every string into []byte
	data := readFile("chatlog_test.log")
	for n := 0; n < b.N; n++ {
		msg := bytes.Split(data, []byte{13, 10}) // []byte("\r\n")
		for _, v := range msg {
			if len(v) == 0 {
				continue
			}
			_, err := parseIRCMessageByte(v)
			if err != nil {
				panic(err)
			}
		}
	}
}

func BenchmarkParseIRCMessageNode(b *testing.B) { // stolen from https://github.com/sigkell/irc-message/blob/master/index.js
	data := readFile("chatlog_test.log")
	for n := 0; n < b.N; n++ {
		msg := bytes.Split(data, []byte{13, 10})
		for _, v := range msg {
			if len(v) == 0 {
				continue
			}
			_, err := parseIRCMessage(v)
			if err != nil {
				panic(err)
			}
		}
	}
}

// goos: windows
// goarch: amd64
// pkg: github.com/spddl/go-twitch-ws
// BenchmarkParseIRCMessageString-8            1830          13076542 ns/op
// BenchmarkParseIRCMessageByte-8              1864          12863764 ns/op
// BenchmarkParseIRCMessageNode-8              3676           6455876 ns/op
// PASS
// ok      github.com/spddl/go-twitch-ws   75.057s
