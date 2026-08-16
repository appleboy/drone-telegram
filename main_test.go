package main

import (
	"os"
	"testing"
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
