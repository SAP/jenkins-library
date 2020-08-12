package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAAKaaSRegisterDeliveryTransportCommand(t *testing.T) {

	testCmd := AAKaaSRegisterDeliveryTransportCommand()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "aAKaaSRegisterDeliveryTransport", testCmd.Use, "command name incorrect")

}
