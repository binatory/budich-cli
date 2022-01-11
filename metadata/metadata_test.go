package metadata

import (
	"fmt"
	"github.com/blang/semver/v4"
	"reflect"
	"testing"
)

func TestShowCurrentVersion(t *testing.T) {
	fmt.Printf("METADATA_CURRENT_VERSION: %s\n", VersionRaw)
}

func Test_makeVersion(t *testing.T) {
	type args struct {
		main       string
		prerelease string
		build      string
	}
	tests := []struct {
		name string
		args args
		want semver.Version
	}{
		{"main only", args{main: "1.0.0"}, semver.MustParse("1.0.0")},
		{"main + pr only", args{main: "1.0.0", prerelease: "alpha"}, semver.MustParse("1.0.0-alpha")},
		{"main + build only", args{main: "1.0.0", build: "abcdef"}, semver.MustParse("1.0.0+abcdef")},
		{"main + pr + build", args{main: "1.0.0", prerelease: "alpha", build: "abcdef"}, semver.MustParse("1.0.0-alpha+abcdef")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeVersion(tt.args.main, tt.args.prerelease, tt.args.build); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
