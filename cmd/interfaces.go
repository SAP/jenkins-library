package cmd

type execRunner interface {
	RunExecutable(e string, p ...string) error
	Dir(d string)
}
