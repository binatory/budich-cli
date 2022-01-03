package utils

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
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

func TestWrapLongRunningFunc(t *testing.T) {
	duration := 200 * time.Millisecond
	expected := errors.New("expected")
	f := func() error {
		<-time.After(duration)
		return expected
	}

	c1 := make(chan error, 1)
	start := time.Now()
	got := <-WrapLongRunningFunc(f, c1)
	require.Equal(t, expected, got)
	require.True(t, time.Now().Sub(start) >= duration)
	require.Equal(t, expected, <-c1)
}
