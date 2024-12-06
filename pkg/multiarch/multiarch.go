package multiarch

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

var knownGoos = []string{"aix", "android", "darwin", "dragonfly", "freebsd", "hurd", "illumos", "ios", "js", "linux", "nacl", "netbsd", "openbsd", "plan9", "solaris", "windows", "zos"}
var knownGoarch = []string{"386", "amd64", "amd64p32", "arm", "arm64", "arm64be", "armbe", "loong64", "mips", "mips64", "mips64le", "mips64p32", "mips64p32le", "mipsle", "ppc", "ppc64", "ppc64le", "riscv", "riscv64", "s390", "s390x", "sparc", "sparc64", "wasm"}

// Platform .
type Platform struct {
	OS      string
	Arch    string
	Variant string
}

// ToString returns a string representation of the platform
func (p Platform) ToString() string {
	if len(p.Variant) > 0 {
		return fmt.Sprintf("%s/%s/%s", p.OS, p.Arch, p.Variant)
	}

	return fmt.Sprintf("%s/%s", p.OS, p.Arch)
}

// ParsePlatformString parses the given string and returns a platform obj
func ParsePlatformString(s string) (Platform, error) {
	r := regexp.MustCompile(`(?P<os>[^,/]+)[,/](?P<arch>[^,/]+)(?:[,/](?P<variant>[^,/]+))?`)

	matches := r.FindStringSubmatch(strings.ToLower(s))

	if len(matches) < 2 {
		return Platform{}, fmt.Errorf("unable to parse platform '%s'", s)
	}

	p := Platform{}

	p.OS = strings.Trim(matches[1], " ")

	if !slices.Contains(knownGoos, p.OS) {
		log.Entry().Warningf("OS '%s' is unknown to us", p.OS)
	}

	p.Arch = strings.Trim(matches[2], " ")

	if !slices.Contains(knownGoarch, p.Arch) {
		log.Entry().Warningf("Architecture '%s' is unknown to us", p.Arch)
	}

	p.Variant = strings.Trim(matches[3], " ")

	return p, nil
}

// ParsePlatformStrings parses the given slice of strings and returns a slice with platform objects
func ParsePlatformStrings(ss []string) ([]Platform, error) {
	pp := []Platform{}

	for _, s := range ss {
		if p, err := ParsePlatformString(s); err == nil {
			pp = append(pp, p)
		} else {
			return nil, err
		}
	}

	return pp, nil
}
