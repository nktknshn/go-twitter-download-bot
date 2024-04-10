package twitter

import (
	"bytes"
	"encoding/json"
)

func JsonDecodeWithNumberString(data string, v any) error {
	return JsonDecodeWithNumberBytes([]byte(data), &v)
}

func JsonDecodeWithNumberBytes(data []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	return d.Decode(&v)
}
