// +build windows linux js,wasm

package twitch

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
)

// go test -bench=. -bench ^(BenchmarkRead.*)$
// goos: windows
// goarch: amd64
// pkg: github.com/spddl/go-twitch-ws
// BenchmarkReadFrom-8               329046              3560 ns/op
// BenchmarkReadNewScanner-8         353674              3371 ns/op
// BenchmarkReadIOUtil-8             435808              2764 ns/op
// PASS
// ok      github.com/spddl/go-twitch-ws   3.801s

func BenchmarkReadFrom(b *testing.B) {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile) // https://ispycode.com/GO/Logging/Setting-output-flags

	f, err := os.Open("chatlog_test.log")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	ioReader := bufio.NewReader(f)

	for i := 0; i < b.N; i++ {
		ioReader.Reset(f)

		bp := bpoolGet()

		_, err := bp.ReadFrom(ioReader)
		if err != nil {
			log.Println(err)
			return
		}

		msg := bytes.Split(bp.Bytes(), []byte{13, 10}) // "\r\n"
		for _, value := range msg {
			// client.parser(value)
			_ = value
		}

		bpoolPut(bp)
	}
}

// stolen from https://github.com/nhooyr/websocket/blob/master/internal/bpool/bpool.go
var bpool sync.Pool // https://golang.org/pkg/sync/#Pool

// Get returns a buffer from the pool or creates a new one if
// the pool is empty.
func bpoolGet() *bytes.Buffer {
	b := bpool.Get()
	if b == nil {
		return &bytes.Buffer{}
	}
	return b.(*bytes.Buffer)
}

// Put returns a buffer into the pool.
func bpoolPut(b *bytes.Buffer) {
	b.Reset()
	bpool.Put(b)
}

func BenchmarkReadNewScanner(b *testing.B) { // https://golang.org/pkg/bufio/#NewScanner
	f, err := os.Open("chatlog_test.log")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	ioReader := bufio.NewReader(f)

	for i := 0; i < b.N; i++ {
		ioReader.Reset(f)
		scanner := bufio.NewScanner(ioReader)
		for scanner.Scan() {
			// client.parser(scanner.Bytes())
		}

		err := scanner.Err()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func BenchmarkReadIOUtil(b *testing.B) { // https://golang.org/src/io/ioutil/ioutil.go?s=1186:1227#L18
	f, err := os.Open("chatlog_test.log")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	ioReader := bufio.NewReader(f)

	for i := 0; i < b.N; i++ {
		ioReader.Reset(f)
		var buf bytes.Buffer
		buf.Grow(bytes.MinRead) // 512

		_, err := buf.ReadFrom(ioReader)
		if err != nil {
			fmt.Println(err)
		}

		msg := bytes.Split(buf.Bytes(), []byte{13, 10}) // "\r\n"
		for _, value := range msg {
			// client.parser(value)
			_ = value
		}
	}
}
