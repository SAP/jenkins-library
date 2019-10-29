package cmd

import (
	"fmt"
)

type mockRunner struct {
	dir   []string
	calls []execCall
}

type execCall struct {
	exec   string
	params []string
}

func (m *mockRunner) Dir(d string) {
	m.dir = append(m.dir, d)
}

func (m *mockRunner) RunExecutable(e string, p ...string) error {
	if e == "fail" {
		return fmt.Errorf("error case")
	}
	exec := execCall{exec: e, params: p}
	m.calls = append(m.calls, exec)
	return nil
}