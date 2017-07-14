package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// RemoteValue is a lazy-man's union type to access upper/lower
// parts of a uint16
type RemoteValue uint16

func (r RemoteValue) GetLow() uint8 {
	return uint8(r & 0xFF)
}

func (r RemoteValue) GetHigh() uint8 {
	return uint8(r >> 8)
}

type RemoteCode struct {
	LeadingBit uint // Single bit
	Channel    uint8
	Remote     RemoteValue
	Action     BlindAction
	Checksum   uint8
}

func Deserialize(code string) (*RemoteCode, error) {
	remoteCode := &RemoteCode{}

	leadingBit, err := strconv.ParseUint(reverse(code[0:1]), 2, 0)
	if err != nil {
		return nil, err
	}
	remoteCode.LeadingBit = uint(leadingBit)

	channel, err := strconv.ParseUint(reverse(code[1:9]), 2, 8)
	if err != nil {
		return nil, err
	}
	remoteCode.Channel = uint8(channel)

	remote, err := strconv.ParseUint(reverse(code[9:25]), 2, 16)
	if err != nil {
		return nil, err
	}
	remoteCode.Remote = RemoteValue(remote)

	actionValue, err := strconv.ParseUint(reverse(code[25:33]), 2, 8)
	if err != nil {
		return nil, err
	}
	action, err := ActionFromValue(uint8(actionValue))
	if err != nil {
		return nil, err
	}
	remoteCode.Action = action

	checksum, err := strconv.ParseUint(reverse(code[33:41]), 2, 8)
	if err != nil {
		return nil, err
	}
	remoteCode.Checksum = uint8(checksum)

	// Sanity Check.
	err = remoteCode.Check()
	if err != nil {
		return nil, err
	}

	return remoteCode, nil
}

// Sanity Check a Remote Code
func (r *RemoteCode) Check() error {
	// Check the fixed regions.
	if r.LeadingBit != 0 {
		return fmt.Errorf("LeadingBit Value sanity check failed")
	}

	return nil
}

func (r *RemoteCode) GuessChecksum() uint8 {
	return r.Channel + r.Remote.GetHigh() + r.Remote.GetLow() + r.Action.Value + 3
}

func (r *RemoteCode) SetChannel(channel uint) {
	// Determine the difference between our channel and the requested channel.
	// diff := channel - r.Channel

	// r.Channel = (r.Channel + diff) % (1 << 8)
	// r.ChannelF = (r.ChannelF + diff) % (1 << 6)

	// if r.Channel < 0 {
	// 	panic("Channel < 0")
	// }

	// if r.ChannelF < 0 {
	// 	r.ChannelF += (1 << 6)
	// }

	// if err := r.Check(); err != nil {
	// 	panic(err)
	// }
}

func (r *RemoteCode) SetAction(action BlindAction) {
	// Determine the difference between our action and the requested action.
	// diff := action - r.Action
	// aFFisFlipped := r.ActionF != r.ActionFF

	// r.Action = (r.Action + diff) % (1 << 2)
	// r.ChannelF = (r.ChannelF + diff) % (1 << 6)

	// if r.Action < 0 {
	// 	panic("Action < 0")
	// }

	// if r.ChannelF < 0 {
	// 	r.ChannelF += (1 << 6)
	// }

	// // Calculate the actionF checksum:
	// if r.Action == 3 {
	// 	r.ActionF = 0
	// } else {
	// 	r.ActionF = 1
	// }

	// // Calculate the actionFF checksum:
	// if aFFisFlipped {
	// 	r.ActionFF = (r.ActionF + 1) % 2
	// } else {
	// 	r.ActionFF = r.ActionF
	// }

	// if err := r.Check(); err != nil {
	// 	panic(err)
	// }

}

type sortableField struct {
	valField  reflect.Value
	typeField reflect.StructField
	tag       RFGenTag
}
type ByTagPos []sortableField

func (b ByTagPos) Len() int           { return len(b) }
func (b ByTagPos) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByTagPos) Less(i, j int) bool { return b[i].tag.pos < b[j].tag.pos }

func (r *RemoteCode) Serialize() string {

	bits := ""
	bits += rightPad(reverse(strconv.FormatUint(uint64(r.LeadingBit), 2)), 1, "0")
	bits += rightPad(reverse(strconv.FormatUint(uint64(r.Channel), 2)), 8, "0")
	bits += rightPad(reverse(strconv.FormatUint(uint64(r.Remote), 2)), 16, "0")
	bits += rightPad(reverse(strconv.FormatUint(uint64(r.Action.Value), 2)), 8, "0")
	bits += rightPad(reverse(strconv.FormatUint(uint64(r.Checksum), 2)), 8, "0")
	return bits

}

type RFGenTag struct {
	pos  int
	bits int
}

func ReadTag(tag string) RFGenTag {
	rfgenTag := RFGenTag{}
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		splitPart := strings.Split(part, "=")
		if splitPart[0] == "pos" {
			val, err := strconv.ParseInt(splitPart[1], 10, 0)
			if err != nil {
				panic(err)
			}

			rfgenTag.pos = int(val)
		} else if splitPart[0] == "bits" {
			val, err := strconv.ParseInt(splitPart[1], 10, 0)
			if err != nil {
				panic(err)
			}
			rfgenTag.bits = int(val)
		}
	}
	return rfgenTag
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func rightPad(s string, n int, with string) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(with, n-len(s))
}
