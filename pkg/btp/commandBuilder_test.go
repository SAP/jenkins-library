package btp

import (
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
		assert.Equal(t, "btp list", cmd)
	})

	t.Run("builds with action and target", func(t *testing.T) {
		builder := NewBTPCommandBuilder().WithAction("get").WithTarget("subaccount")
		cmd, err := builder.Build()
		assert.NoError(t, err)
		assert.Equal(t, "btp get subaccount", cmd)
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
		assert.Equal(t, "btp --global delete service-instance --force my-instance", cmd)
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
		assert.Equal(t, "btp --global --json create service-instance --name test --plan standard", cmd)
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
		assert.NoError(t, err)
		assert.Contains(t, cmd, "btp get subaccount")
		assert.Contains(t, cmd, "--subaccount 123")
		assert.Contains(t, cmd, "--id my-id")
		assert.Contains(t, cmd, "--name my-name")
		assert.Contains(t, cmd, "--show-parameters true")
		assert.Contains(t, cmd, "--data-center eu10")
		assert.Contains(t, cmd, "--service hana")
		assert.Contains(t, cmd, "--plan plan-123")
		assert.Contains(t, cmd, "--plan-name plan-name")
		assert.Contains(t, cmd, "--offering-name offering")
		assert.Contains(t, cmd, "--parameters {\"foo\":\"bar\"}")
		assert.Contains(t, cmd, "--labels {\"env\":\"prod\"}")
		assert.Contains(t, cmd, "--confirm true")
		assert.Contains(t, cmd, "--binding bind1")
		assert.Contains(t, cmd, "--instance-name svc1")
		assert.Contains(t, cmd, "--service-instancee svc-id")
		assert.Contains(t, cmd, "--url https://example.com")
		assert.Contains(t, cmd, "--subdomain mydomain")
		assert.Contains(t, cmd, "--user user")
		assert.Contains(t, cmd, "--password pass")
		assert.Contains(t, cmd, "--format json")
		assert.Contains(t, cmd, "--verbose")
		assert.Contains(t, cmd, "--idp tenant-1")
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
