package main

import (
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/urfave/cli"
)

func TestWidgetWithEmptyActionInputs(t *testing.T) {
	if os.Getenv("WIDGET_CLI_HELPER") == "1" {
		os.Args = []string{"drone-telegram"}
		main()
		return
	}
	for _, format := range []string{"", "HTML"} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestWidgetWithEmptyActionInputs$")
		cmd.Env = []string{"WIDGET_CLI_HELPER=1", "GITHUB=true", "INPUT_TOKEN=fake", "INPUT_TO=42", "INPUT_WIDGET_FILE=/nonexistent/widget.json", "INPUT_FORMAT=" + format}
		for _, key := range []string{"MESSAGE", "MESSAGE_FILE", "RICH_MESSAGE_FILE", "PHOTO", "DOCUMENT", "STICKER", "AUDIO", "VOICE", "VIDEO", "LOCATION", "VENUE", "DEBUG", "DISABLE_WEB_PAGE_PREVIEW", "DISABLE_NOTIFICATION", "MESSAGE_THREAD_ID", "SOCKS5"} {
			cmd.Env = append(cmd.Env, "INPUT_"+key+"=")
		}
		output, err := cmd.CombinedOutput()
		want := "cannot read widget_file"
		if format != "" {
			want = "cannot be combined"
		}
		if err == nil || !strings.Contains(string(output), want) {
			t.Fatalf("format %q: err=%v output=%s", format, err, output)
		}
	}
}

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
