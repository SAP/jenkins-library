package multiarch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlatformToString(t *testing.T) {
	tt := []struct {
		uut    Platform
		expect string
	}{
		{
			uut: Platform{
				OS:   "linux",
				Arch: "arm64",
			},
			expect: "linux/arm64",
		},
		{
			uut: Platform{
				OS:      "linux",
				Arch:    "arm64",
				Variant: "v8",
			},
			expect: "linux/arm64/v8",
		},
	}

	for _, test := range tt {
		t.Run(test.expect, func(t *testing.T) {
			assert.Equal(t, test.expect, test.uut.ToString())
		})
	}
}

func TestParsePlatformString(t *testing.T) {
	tt := []struct {
		description string
		input       string
		expect      Platform
		expectError string
	}{
		{
			description: "format used by golangBuild - only os + arch",
			input:       "linux,amd64",
			expect: Platform{
				OS:   "linux",
				Arch: "amd64",
			},
		},
		{
			description: "format used by kanikoExecute - os/arch/variant",
			input:       "linux/amd64/v8",
			expect: Platform{
				OS:      "linux",
				Arch:    "amd64",
				Variant: "v8",
			},
		},
		{
			description: "sth in between - os,arch,variant",
			input:       "linux,amd64,v8",
			expect: Platform{
				OS:      "linux",
				Arch:    "amd64",
				Variant: "v8",
			},
		},
		{
			description: "should be case insensitive",
			input:       "LINUX/AMD64/V8",
			expect: Platform{
				OS:      "linux",
				Arch:    "amd64",
				Variant: "v8",
			},
		},
		{
			description: "whitespaces shall be trimmed",
			input:       "    linux/ amd64  / v8",
			expect: Platform{
				OS:      "linux",
				Arch:    "amd64",
				Variant: "v8",
			},
		},
		{
			description: "reads unsupported values",
			input:       "myfancyos/runningonafancyarchitecture",
			expect: Platform{
				OS:   "myfancyos",
				Arch: "runningonafancyarchitecture",
			},
		},
		{
			description: "os + arch are required, variant is optional",
			input:       "linux",
			expectError: "unable to parse platform 'linux'",
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			p, err := ParsePlatformString(test.input)

			if len(test.expectError) > 0 {
				assert.EqualError(t, err, test.expectError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.expect, p)
		})
	}
}

func TestParsePlatformStringStrings(t *testing.T) {
	tt := []struct {
		description string
		inputs      []string
		expect      []Platform
		expectError string
	}{
		{
			description: "format used by golangBuild - only os + arch",
			inputs:      []string{"linux,amd64", "windows,amd64"},
			expect: []Platform{
				Platform{
					OS:   "linux",
					Arch: "amd64",
				},
				Platform{
					OS:   "windows",
					Arch: "amd64",
				},
			},
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			p, err := ParsePlatformStrings(test.inputs)

			if len(test.expectError) > 0 {
				assert.EqualError(t, err, test.expectError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.expect, p)
		})
	}
}
