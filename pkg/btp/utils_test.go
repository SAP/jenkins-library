package btp

import "testing"

func TestNewBTPUtils(t *testing.T) {
	t.Run("NewBTPUtils sets Exec correctly", func(t *testing.T) {
		m := &BtpExecutorMock{}
		btpUtils := NewBTPUtils(m)

		if btpUtils.Exec != m {
			t.Errorf("Expected Exec to be set to the provided ExecRunner, got %v", btpUtils.Exec)
		}
	})
}
