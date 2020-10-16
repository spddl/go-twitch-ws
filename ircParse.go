// stolen from https://github.com/sigkell/irc-message/blob/master/index.js

package twitch

import (
	"bytes"
)

type IRCMessage struct {
	Raw     []byte
	Tags    map[string][]byte
	Command []byte
	Params  [][]byte
	Prefix  []byte
}

func parseIRCMessage(data []byte) (*IRCMessage, error) {
	message := IRCMessage{
		Raw:    data,
		Tags:   map[string][]byte{},
		Params: [][]byte{},
	}

	// position and nextspace are used by the parser as a reference.
	var position int
	var nextspace int

	// The first thing we check for is IRCv3.2 message tags.
	// http://ircv3.atheme.org/specification/message-tags-3.2

	if data[0] == 64 {
		var nextspace = bytes.Index(data, []byte{32})
		if nextspace == -1 {
			// Malformed IRC message.
			return &message, nil
		}

		// Tags are split by a semi colon.
		var rawTags = bytes.Split(data[1:nextspace], []byte{59})

		for i := 0; i < len(rawTags); i++ {
			// Tags delimited by an equals sign are key=value tags.
			// If there's no equals, we assign the tag a value of true.
			var tag = rawTags[i]
			var pair = bytes.Split(tag, []byte{61})

			if len(pair[1]) == 0 {
				// message.Tags[string(pair[0])] = []byte{116, 114, 117, 101} // true string
				message.Tags[string(pair[0])] = []byte{} // empty string
			} else {
				message.Tags[string(pair[0])] = pair[1]
			}
		}

		position = nextspace + 1
	}

	// Skip any trailing whitespace.
	for data[position] == 32 {
		position++
	}

	// Extract the message's prefix if present. Prefixes are prepended
	// with a colon.
	if data[position] == 58 {
		nextspace = bytes.Index(data[position:], []byte{32})
		if nextspace != -1 {
			nextspace += position
		}

		// If there's nothing after the prefix, deem this message to be
		// malformed.
		if nextspace == -1 {
			// Malformed IRC message.
			return &message, nil
		}

		message.Prefix = data[position+1 : nextspace]
		position = nextspace + 1

		// Skip any trailing whitespace.
		for data[position] == 32 {
			position++
		}
	}

	nextspace = bytes.Index(data[position:], []byte{32})
	if nextspace != -1 {
		nextspace += position
	}

	// If there's no more whitespace left, extract everything from the
	// current position to the end of the string as the command.
	if nextspace == -1 {
		if len(data) > position {
			message.Command = data[position:]
			return &message, nil
		}

		return nil, nil
	}

	// Else, the command is the current position up to the next space. After
	// that, we expect some parameters.
	message.Command = data[position:nextspace]

	position = nextspace + 1

	// Skip any trailing whitespace.
	for data[position] == 32 {
		position++
	}

	for position < len(data) {
		nextspace = bytes.Index(data[position:], []byte{32})
		if nextspace != -1 {
			nextspace += position
		}

		// If the character is a colon, we've got a trailing parameter.
		// At this point, there are no extra params, so we push everything
		// from after the colon to the end of the string, to the params array
		// and break out of the loop.
		if data[position] == 58 {
			message.Params = append(message.Params, data[position+1:])
			break
		}

		// If we still have some whitespace...
		if nextspace != -1 {
			// Push whatever's between the current position and the next
			// space to the params array.
			message.Params = append(message.Params, data[position:nextspace])
			position = nextspace + 1

			// Skip any trailing whitespace and continue looping.
			for data[position] == 32 {
				position++
			}
			continue
		}

		// If we don't have any more whitespace and the param isn't trailing,
		// push everything remaining to the params array.
		if nextspace == -1 {
			message.Params = append(message.Params, data[position:])
			break
		}
	}

	return &message, nil
}
