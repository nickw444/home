package main

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type RemoteCode struct {
	Channel int `rfgen:"pos=1,bits=6"`
	Remote  int `rfgen:"pos=7,bits=15"`
	Action  int `rfgen:"pos=25,bits=2"`

	ActionF  int `rfgen:"pos=32,bits=1"`
	ChannelF int `rfgen:"pos=33,bits=6"`
	MF       int `rfgen:"pos=39,bits=1"`
	ActionFF int `rfgen:"pos=40,bits=1"`

	Fixed1 int `rfgen:"pos=0,bits=1"`
	Fixed2 int `rfgen:"pos=22,bits=3"`
	Fixed3 int `rfgen:"pos=27,bits=5"`
}

func getActionName(action int) string {
	if action == 0 {
		return "DOWN"
	} else if action == 1 {
		return "STOP"
	} else if action == 2 {
		return "UP"
	} else if action == 3 {
		return "PAIR"
	}
	return ""
}

func getActionValue(action string) (int, error) {
	if action == "DOWN" {
		return 0, nil
	} else if action == "STOP" {
		return 1, nil
	} else if action == "UP" {
		return 2, nil
	} else if action == "PAIR" {
		return 3, nil
	}
	return -1, fmt.Errorf("Unknown action: %s", action)
}

func Deserialize(code string) (remoteCode *RemoteCode, err error) {
	remoteCode = &RemoteCode{}
	val := reflect.ValueOf(remoteCode).Elem()

	for i := 0; i < val.NumField(); i++ {
		valField := val.Field(i)
		typeField := val.Type().Field(i)

		strTag, ok := typeField.Tag.Lookup("rfgen")
		if ok {
			tag := ReadTag(strTag)
			if valField.CanSet() {
				bits := code[tag.pos : tag.bits+tag.pos]
				bits = reverse(bits)
				readVal, err := strconv.ParseInt(bits, 2, 0)
				if err != nil {
					return nil, err
				}
				valField.SetInt(readVal)
			}
		}

	}

	// Sanity Check.
	err = remoteCode.Check()
	if err != nil {
		return nil, err
	}

	return
}

// Sanity Check a Remote Code
func (r *RemoteCode) Check() error {
	// Check the fixed regions.
	if r.Fixed1 != 0 {
		return fmt.Errorf("Fixed1 Value sanity check failed")
	}
	if r.Fixed2 != 6 {
		return fmt.Errorf("Fixed2 Value sanity check failed")
	}
	if r.Fixed3 != 31 {
		return fmt.Errorf("Fixed3 Value sanity check failed")
	}

	if (r.Action == 3 && r.ActionF != 0) || (r.Action != 3 && r.ActionF != 1) {
		return fmt.Errorf("ActionF value sanity check failed")
	}

	return nil
}

func (r *RemoteCode) SetChannel(channel int) {
	// Determine the difference between our channel and the requested channel.
	diff := channel - r.Channel
	mod := 1 << 6

	r.Channel = (r.Channel + diff) % mod
	r.ChannelF = (r.ChannelF + diff) % mod

	if r.Channel < 0 {
		panic("Channel < 0")
	}

	if r.ChannelF < 0 {
		r.ChannelF += mod
	}

	if err := r.Check(); err != nil {
		panic(err)
	}
}

func (r *RemoteCode) SetAction(action int) {
	// Determine the difference between our action and the requested action.
	diff := action - r.Action
	aFFisFlipped := r.ActionF != r.ActionFF

	r.Action = (r.Action + diff) % (1 << 2)
	r.ChannelF = (r.ChannelF + diff) % (1 << 6)

	if r.Action < 0 {
		panic("Action < 0")
	}

	if r.ChannelF < 0 {
		r.ChannelF += (1 << 6)
	}

	// Calculate the actionF checksum:
	if r.Action == 3 {
		r.ActionF = 0
	} else {
		r.ActionF = 1
	}

	// Calculate the actionFF checksum:
	if aFFisFlipped {
		r.ActionFF = (r.ActionF + 1) % 2
	} else {
		r.ActionFF = r.ActionF
	}

	if err := r.Check(); err != nil {
		panic(err)
	}

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
	val := reflect.ValueOf(r).Elem()
	fields := []sortableField{}

	for i := 0; i < val.NumField(); i++ {
		valField := val.Field(i)
		typeField := val.Type().Field(i)

		strTag, ok := typeField.Tag.Lookup("rfgen")
		if ok {
			tag := ReadTag(strTag)
			fields = append(fields, sortableField{
				valField:  valField,
				typeField: typeField,
				tag:       tag,
			})
		}
	}

	sort.Sort(ByTagPos(fields))

	data := ""
	for _, field := range fields {
		val := field.valField.Int()
		bits := reverse(strconv.FormatInt(val, 2))
		data += rightPad(bits, field.tag.bits, "0")
	}

	return data
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
