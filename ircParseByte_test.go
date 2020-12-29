// +build windows linux js,wasm

// stolen from https://github.com/gempir/go-twitch-irc/blob/master/irc.go
// and convert it all too byte arrays

package twitch

import (
	"bytes"
	"fmt"
)

var tagEscapeCharactersByte = []struct {
	from []byte
	to   []byte
}{
	{[]byte{92, 115}, []byte{58}}, // {`\s`, ` `},
	{[]byte{92, 110}, []byte{}},   // {`\n`, ``},
	{[]byte{92, 114}, []byte{}},   // {`\r`, ``},
	{[]byte{92, 58}, []byte{59}},  // {`\:`, `;`},
	{[]byte{92, 92}, []byte{92}},  // {`\\`, `\`},
}

type IRCMessageSourceByte struct {
	Nickname []byte
	Username []byte
	Host     []byte
}

type IRCMessageByte struct {
	Tags    map[string][]byte
	Source  IRCMessageSourceByte
	Command []byte
	Params  [][]byte
}

func parseIRCMessageByte(line []byte) (*IRCMessageByte, error) {
	message := IRCMessageByte{
		Tags:   map[string][]byte{},
		Params: [][]byte{},
	}

	split := bytes.Split(line, []byte{32}) // space
	index := 0

	if bytes.HasPrefix(split[index], []byte{64}) { // @
		message.Tags = parseIRCTags(split[index])
		index++
	}

	if index >= len(split) {
		return &message, fmt.Errorf("parseIRCMessage: partial message")
	}

	if bytes.HasPrefix(split[index], []byte{58}) { // :
		message.Source = *parseIRCMessageSource(split[index])
		index++
	}

	if index >= len(split) {
		return &message, fmt.Errorf("parseIRCMessage: no command")
	}

	message.Command = split[index]
	index++

	if index >= len(split) {
		return &message, nil
	}

	var params [][]byte
	for i, v := range split[index:] {
		if bytes.HasPrefix(v, []byte{58}) { // :
			v = bytes.Join(split[index+i:], []byte{32}) // space
			v = bytes.TrimPrefix(v, []byte{58})         // :
			params = append(params, v)
			break
		}

		params = append(params, v)
	}

	message.Params = params

	return &message, nil
}

func parseIRCTags(rawTags []byte) map[string][]byte {
	tags := map[string][]byte{}

	rawTags = bytes.TrimPrefix(rawTags, []byte{64}) // @

	for _, tag := range bytes.Split(rawTags, []byte{59}) { // ;
		pair := bytes.SplitN(tag, []byte{61}, 2) // =
		key := string(pair[0])

		var value []byte
		if len(pair) == 2 {
			value = parseIRCTagValue(pair[1])
		}
		tags[key] = value
	}

	return tags
}

func parseIRCTagValue(rawValue []byte) []byte {
	for _, escape := range tagEscapeCharactersByte {
		rawValue = bytes.ReplaceAll(rawValue, escape.from, escape.to)
	}

	rawValue = bytes.TrimSuffix(rawValue, []byte{92}) // "\\"

	// Some Twitch values can end with a trailing \s
	// Example: "system-msg=An\sanonymous\suser\sgifted\sa\sTier\s1\ssub\sto\sTenureCalculator!\s"
	rawValue = bytes.TrimSpace(rawValue)

	return rawValue
}

func regexSplit(r rune) bool {
	return r == '!' || r == '@'
}

func parseIRCMessageSource(rawSource []byte) *IRCMessageSourceByte {
	var source IRCMessageSourceByte

	rawSource = bytes.TrimPrefix(rawSource, []byte{58}) // :

	split := bytes.FieldsFunc(rawSource, regexSplit)

	if len(split) == 0 {
		return &source
	}

	switch len(split) {
	case 1:
		source.Host = split[0]
	case 2:
		// Getting 2 items extremely rare, but does happen sometimes.
		// https://github.com/gempir/go-twitch-irc/issues/109
		source.Nickname = split[0]
		source.Host = split[1]
	default:
		source.Nickname = split[0]
		source.Username = split[1]
		source.Host = split[2]
	}

	return &source
}
