package cmd

type execRunner interface {
	RunExecutable(e string, p ...string) error
	Dir(d string)
}

type shellRunner interface {
	RunShell(s string, c string) error
	Dir(d string)
}
