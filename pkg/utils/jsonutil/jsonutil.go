package jsonutil

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2/log"
)

func ToJsonStr(v any) string {
	var b, err = json.Marshal(v)
	if err != nil {
		log.Error(err)
	}
	return string(b)
}

func ToJsonBytes(v any) []byte {
	var b, err = json.Marshal(v)
	if err != nil {
		log.Error(err)
	}
	return b
}
