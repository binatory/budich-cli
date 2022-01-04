package utils

import (
	"github.com/stretchr/testify/require"
	"math"
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

func TestSecondsToDuration(t *testing.T) {
	type args struct {
		seconds int64
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{"happy case", args{10}, 10 * time.Second},
		{"minDuration (overflowed)", args{math.MinInt64}, 0},
		{"maxDuration (overflowed)", args{math.MaxInt64}, -1 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SecondsToDuration(tt.args.seconds); got != tt.want {
				t.Errorf("SecondsToDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
