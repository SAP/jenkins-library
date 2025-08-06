package orchestrator

import (
	"os"
	"strings"
)

func resetEnv(e []string) {
	for _, val := range e {
		tmp := strings.Split(val, "=")
		os.Setenv(tmp[0], tmp[1])
	}
}
