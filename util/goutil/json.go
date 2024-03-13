package goutil

import "encoding/json"

func JsonValue[T any](bs []byte) (T, error) {
	var val T
	err := json.Unmarshal(bs, &val)
	return val, err
}

func JsonString(item interface{}) string {
	bs, _ := json.Marshal(item)
	return string(bs)
}
