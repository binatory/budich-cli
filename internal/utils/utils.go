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

func WrapLongRunningFunc(f func() error, channels ...chan error) chan error {
	ret := make(chan error)
	channels = append(channels, ret)

	go func() {
		err := f()
		for _, c := range channels {
			c <- err
			close(c)
		}
	}()

	return ret
}
