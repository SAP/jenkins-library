package btp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBTPCommandBuilder_Build(t *testing.T) {
	t.Run("fails if action is not set", func(t *testing.T) {
		builder := NewBTPCommandBuilder()
		cmd, err := builder.Build()
		assert.Error(t, err)
		assert.Empty(t, cmd)
	})

	t.Run("builds with only action", func(t *testing.T) {
		builder := NewBTPCommandBuilder().WithAction("list")
		cmd, err := builder.Build()
		assert.NoError(t, err)
		assert.Equal(t, "btp list", strings.Join(cmd, " "))
	})

	t.Run("builds with action and target", func(t *testing.T) {
		builder := NewBTPCommandBuilder().WithAction("get").WithTarget("subaccount")
		cmd, err := builder.Build()
		assert.NoError(t, err)
		assert.Equal(t, "btp get subaccount", strings.Join(cmd, " "))
	})

	t.Run("builds with action, target, options, and params", func(t *testing.T) {
		builder := NewBTPCommandBuilder().
			WithOption("--global").
			WithAction("delete").
			WithTarget("service-instance").
			WithParam("--force").
			WithParam("my-instance")
		cmd, err := builder.Build()
		assert.NoError(t, err)
		assert.Equal(t, "btp --global delete service-instance --force my-instance", strings.Join(cmd, " "))
	})

	t.Run("builds with multiple options and params", func(t *testing.T) {
		builder := NewBTPCommandBuilder().
			WithOption("--global").
			WithOption("--json").
			WithAction("create").
			WithTarget("service-instance").
			WithParam("--name").
			WithParam("test").
			WithParam("--plan").
			WithParam("standard")
		cmd, err := builder.Build()
		assert.NoError(t, err)
		assert.Equal(t, "btp --global --json create service-instance --name test --plan standard", strings.Join(cmd, " "))
	})

	t.Run("builds with specific parameter methods", func(t *testing.T) {
		builder := NewBTPCommandBuilder().
			WithAction("get").
			WithTarget("subaccount").
			WithSubAccount("123").
			WithID("my-id").
			WithName("my-name").
			SetShowParameters(true).
			WithDataCenter("eu10").
			WithService("hana").
			WithPlanID("plan-123").
			WithPlanName("plan-name").
			WithOfferingName("offering").
			WithParameters("{\"foo\":\"bar\"}").
			WithLabels("{\"env\":\"prod\"}").
			SetConfirm(true).
			WithBindingName("bind1").
			WithServiceInstanceName("svc1").
			WithServiceInstanceID("svc-id").
			WithURL("https://example.com").
			WithSubdomain("mydomain").
			WithUser("user").
			WithPassword("pass").
			WithFormat("json").
			WithVerbose().
			WithTenant("tenant-1")

		cmd, err := builder.Build()

		cmdString := strings.Join(cmd, " ")
		assert.NoError(t, err)
		assert.Contains(t, cmdString, "btp get subaccount")
		assert.Contains(t, cmdString, "--subaccount 123")
		assert.Contains(t, cmdString, "--id my-id")
		assert.Contains(t, cmdString, "--name my-name")
		assert.Contains(t, cmdString, "--show-parameters true")
		assert.Contains(t, cmdString, "--data-center eu10")
		assert.Contains(t, cmdString, "--service hana")
		assert.Contains(t, cmdString, "--plan plan-123")
		assert.Contains(t, cmdString, "--plan-name plan-name")
		assert.Contains(t, cmdString, "--offering-name offering")
		assert.Contains(t, cmdString, "--parameters {\"foo\":\"bar\"}")
		assert.Contains(t, cmdString, "--labels {\"env\":\"prod\"}")
		assert.Contains(t, cmdString, "--confirm true")
		assert.Contains(t, cmdString, "--binding bind1")
		assert.Contains(t, cmdString, "--instance-name svc1")
		assert.Contains(t, cmdString, "--service-instancee svc-id")
		assert.Contains(t, cmdString, "--url https://example.com")
		assert.Contains(t, cmdString, "--subdomain mydomain")
		assert.Contains(t, cmdString, "--user user")
		assert.Contains(t, cmdString, "--password pass")
		assert.Contains(t, cmdString, "--format json")
		assert.Contains(t, cmdString, "--verbose")
		assert.Contains(t, cmdString, "--idp tenant-1")
	})

	t.Run("reset clears all fields", func(t *testing.T) {
		builder := NewBTPCommandBuilder().
			WithOption("--global").
			WithAction("delete").
			WithTarget("service-instance").
			WithParam("--force")
		builder.Reset()
		cmd, err := builder.Build()
		assert.Error(t, err)
		assert.Empty(t, cmd)
	})
}
