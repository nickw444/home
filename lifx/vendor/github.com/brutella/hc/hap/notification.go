package hap

import (
	"bytes"
	"encoding/json"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/hap/data"
	"io/ioutil"
	"net/http"
	"strings"
)

// NewCharacteristicNotification returns an notification response for a characteristic from an accessory.
func NewCharacteristicNotification(a *accessory.Accessory, c *characteristic.Characteristic) (*http.Response, error) {
	body, err := Body(a, c)
	if err != nil {
		return nil, err
	}

	return NewNotification(body), nil
}

// NewNotification returns a notification response with a specific body content.
func NewNotification(body *bytes.Buffer) *http.Response {
	resp := new(http.Response)
	resp.Status = "200 OK"
	resp.StatusCode = http.StatusOK
	resp.ProtoMajor = 1
	resp.ProtoMinor = 0
	resp.Body = ioutil.NopCloser(body)
	resp.ContentLength = int64(body.Len())
	resp.Header = map[string][]string{}
	resp.Header.Set("Content-Type", HTTPContentTypeHAPJson)

	// Will be ignored unfortunately and won't be fixed https://github.com/golang/go/issues/9304
	// Make sure to call FixProtocolSpecifier() instead
	resp.Proto = "EVENT/1.0"

	return resp
}

// FixProtocolSpecifier returns bytes where the http protocol specifier "HTTP/1.0" is replaced by "EVENT/1.0" in the argument bytes.
// This fix is necessary because http.Response ignores the Proto field value.
//
// Related to issue: https://github.com/golang/go/issues/9304
func FixProtocolSpecifier(b []byte) []byte {
	return []byte(strings.Replace(string(b), "HTTP/1.0", "EVENT/1.0", 1))
}

// Body returns the json body for an notification response as bytes.
func Body(a *accessory.Accessory, c *characteristic.Characteristic) (*bytes.Buffer, error) {

	ch := data.Characteristic{AccessoryID: a.GetID(), CharacteristicID: c.GetID(), Value: c.Value}
	chars := data.Characteristics{[]data.Characteristic{ch}}
	result, err := json.Marshal(chars)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	b.Write(result)
	return &b, err
}
