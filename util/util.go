package util

import (
	"encoding/json"
	"math/rand"
	"time"
)

func ConvertArray(source interface{}) []interface{} {
	switch source.(type) {
	case []string:
		tmp := source.([]string)
		opt := make([]interface{}, len(tmp))
		for i, obj := range tmp {
			opt[i] = obj
		}
		return opt
	default:
		return nil
	}
}

func GetOneRandom(src interface{}) interface{} {
	array := ConvertArray(src)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	item := array[r.Intn(len(array))]
	return item
}

// only work for string
func GetOneLevelJson(body, key string) string {
	var mapping map[string]string
	json.Unmarshal([]byte(body), &mapping)
	return mapping[key]
}