package main

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/urfave/cli"
)

func TestUnsetEmptyEnv(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		set       bool
		wantExist bool
	}{
		{"empty value removed", "", true, false},
		{"whitespace value removed", "   ", true, false},
		{"non-empty value kept", "12345", true, true},
		{"unset stays unset", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_UNSET_EMPTY_ENV"
			os.Unsetenv(key)
			if tt.set {
				t.Setenv(key, tt.value)
			}

			unsetEmptyEnv(key)

			if _, ok := os.LookupEnv(key); ok != tt.wantExist {
				t.Errorf("LookupEnv(%q) exist = %v, want %v", key, ok, tt.wantExist)
			}
			if tt.wantExist {
				if got := os.Getenv(key); got != tt.value {
					t.Errorf("Getenv(%q) = %q, want %q", key, got, tt.value)
				}
			}
		})
	}
}

func TestWidgetIgnoresOnlyImplicitTextFormat(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		flags := flag.NewFlagSet("widget", flag.ContinueOnError)
		flags.String("widget.file", "/nonexistent/widget.json", "")
		flags.String("token", "fake", "")
		flags.String("format", formatMarkdown, "")
		recipients := cli.StringSlice{"42"}
		flags.Var(&recipients, "to", "")
		args := []string{}
		if explicit {
			args = append(args, "--format", "HTML")
		}
		if err := flags.Parse(args); err != nil {
			t.Fatal(err)
		}
		err := run(cli.NewContext(cli.NewApp(), flags, nil))
		want := "cannot read widget_file"
		if explicit {
			want = "cannot be combined"
		}
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("unexpected validation: %v", err)
		}
	}
}
