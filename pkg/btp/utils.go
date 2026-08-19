package btp

func NewBTPUtils(exec ExecRunner) *BTPUtils {
	b := new(BTPUtils)
	b.Exec = exec
	return b
}

type BTPUtils struct {
	Exec ExecRunner
}
