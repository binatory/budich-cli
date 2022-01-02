package utils

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetMapKeys(t *testing.T) {
	m := map[string]interface{}{
		"zzkey": 1,
		"key1":  "string",
		"key2":  true,
	}
	keys := GetMapKeys(m)
	require.EqualValues(t, []string{"key1", "key2", "zzkey"}, keys)
}
