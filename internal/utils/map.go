package utils

import (
	"reflect"
	"sort"
)

func GetMapKeys(m interface{}) []string {
	t := reflect.TypeOf(m)
	if t.Kind() != reflect.Map || t.Key().Kind() != reflect.String {
		panic("m must be a map with keys of type string")
	}

	var res []string
	for _, keyVal := range reflect.ValueOf(m).MapKeys() {
		res = append(res, keyVal.String())
	}
	sort.Strings(res)
	return res
}
