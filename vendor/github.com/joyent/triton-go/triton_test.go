package triton_test

import (
	"os"
	"testing"

	triton "github.com/joyent/triton-go"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name    string
		varname string
		input   string
		value   string
	}{
		{"Triton", "TRITON_NAME", "NAME", "good"},
		{"SDC", "SDC_NAME", "NAME", "good"},
		{"unrelated", "BAD_NAME", "NAME", ""},
		{"missing", "", "NAME", ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			os.Setenv(test.varname, test.value)
			defer os.Unsetenv(test.varname)

			if val := triton.GetEnv(test.input); val != test.value {
				t.Errorf("expected %s env var to be '%s': got '%s'",
					test.varname, test.value, val)
			}
		})
	}
}
