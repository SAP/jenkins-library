package btp

import (
	"errors"
	"strconv"
)

// Initialize a new builder
func NewBTPCommandBuilder() *BTPCommandBuilder {
	return &BTPCommandBuilder{}
}

// Rest builder
func (b *BTPCommandBuilder) Reset() *BTPCommandBuilder {
	b.options = []string{}
	b.action = ""
	b.target = ""
	b.params = []string{}

	return b
}

// General method to add an option
func (b *BTPCommandBuilder) WithOption(option string) *BTPCommandBuilder {
	b.options = append(b.options, option)
	return b
}

// Method to set action
func (b *BTPCommandBuilder) WithAction(action string) *BTPCommandBuilder {
	b.action = action
	return b
}

// Method to set target (e.g., GROUP/OBJECT)
func (b *BTPCommandBuilder) WithTarget(target string) *BTPCommandBuilder {
	b.target = target
	return b
}

// Method to add additional parameters
func (b *BTPCommandBuilder) WithParam(param string) *BTPCommandBuilder {
	b.params = append(b.params, param)
	return b
}

// Specific method for --subaccount parameter
func (b *BTPCommandBuilder) WithSubAccount(subaccount string) *BTPCommandBuilder {
	b.params = append(b.params, "--subaccount", subaccount)
	return b
}

// Specific method for --labels-filter parameter
func (b *BTPCommandBuilder) WithLabelsFilter(query string) *BTPCommandBuilder {
	b.params = append(b.params, "--labels-filter", query)
	return b
}

// Specific method for --fields-filter parameter
func (b *BTPCommandBuilder) WithFieldsFilter(query string) *BTPCommandBuilder {
	b.params = append(b.params, "--fields-filter", query)
	return b
}

// Specific method for --id parameter
func (b *BTPCommandBuilder) WithID(id string) *BTPCommandBuilder {
	b.params = append(b.params, "--id", id)
	return b
}

// Specific method for --name parameter
func (b *BTPCommandBuilder) WithName(name string) *BTPCommandBuilder {
	b.params = append(b.params, "--name", name)
	return b
}

// Specific method for --show-parameters parameter
func (b *BTPCommandBuilder) SetShowParameters(value bool) *BTPCommandBuilder {
	b.params = append(b.params, "--show-parameters", strconv.FormatBool(value))
	return b
}

// Specific method for --data-center parameter
func (b *BTPCommandBuilder) WithDataCenter(name string) *BTPCommandBuilder {
	b.params = append(b.params, "--data-center", name)
	return b
}

// Specific method for --service parameter
func (b *BTPCommandBuilder) WithService(name string) *BTPCommandBuilder {
	b.params = append(b.params, "--service", name)
	return b
}

// Specific method for --plan parameter
func (b *BTPCommandBuilder) WithPlanID(id string) *BTPCommandBuilder {
	b.params = append(b.params, "--plan", id)
	return b
}

// Specific method for --plan-name parameter
func (b *BTPCommandBuilder) WithPlanName(name string) *BTPCommandBuilder {
	b.params = append(b.params, "--plan-name", name)
	return b
}

// Specific method for --offering-name parameter
func (b *BTPCommandBuilder) WithOfferingName(name string) *BTPCommandBuilder {
	b.params = append(b.params, "--offering-name", name)
	return b
}

// Specific method for --offering-name parameter
func (b *BTPCommandBuilder) WithParameters(json string) *BTPCommandBuilder {
	b.params = append(b.params, "--parameters", json)
	return b
}

// Specific method for --labels parameter
func (b *BTPCommandBuilder) WithLabels(json string) *BTPCommandBuilder {
	b.params = append(b.params, "--labels", json)
	return b
}

// Specific method for --confirm parameter
func (b *BTPCommandBuilder) SetConfirm(value bool) *BTPCommandBuilder {
	b.params = append(b.params, "--confirm", strconv.FormatBool(value))
	return b
}

// Specific method for --binding parameter
func (b *BTPCommandBuilder) WithBindingName(name string) *BTPCommandBuilder {
	b.params = append(b.params, "--binding", name)
	return b
}

// Specific method for --instance-name parameter
func (b *BTPCommandBuilder) WithServiceInstanceName(name string) *BTPCommandBuilder {
	b.params = append(b.params, "--instance-name", name)
	return b
}

// Specific method for --instance-name parameter
func (b *BTPCommandBuilder) WithServiceInstanceID(id string) *BTPCommandBuilder {
	b.params = append(b.params, "--service-instancee", id)
	return b
}

// Specific method for --url parameter
func (b *BTPCommandBuilder) WithURL(url string) *BTPCommandBuilder {
	b.params = append(b.params, "--url", url)
	return b
}

// Specific method for --subdomain parameter
func (b *BTPCommandBuilder) WithSubdomain(globalAccount string) *BTPCommandBuilder {
	b.params = append(b.params, "--subdomain", globalAccount)
	return b
}

// Specific method for --user parameter
func (b *BTPCommandBuilder) WithUser(user string) *BTPCommandBuilder {
	b.params = append(b.params, "--user", user)
	return b
}

// Specific method for --password parameter
func (b *BTPCommandBuilder) WithPassword(password string) *BTPCommandBuilder {
	b.params = append(b.params, "--password", password)
	return b
}

func (b *BTPCommandBuilder) WithConfirm() *BTPCommandBuilder {
	b.params = append(b.params, "--confirm")
	return b
}

// Method to set format of response
func (b *BTPCommandBuilder) WithFormat(format string) *BTPCommandBuilder {
	b.params = append(b.params, "--format", format)
	return b
}

// Method to activate verbose mode
func (b *BTPCommandBuilder) WithVerbose() *BTPCommandBuilder {
	b.params = append(b.params, "--verbose", "true")
	return b
}

// Specific method for --idp parameter
func (b *BTPCommandBuilder) WithIdentityProvider(idProvider string) *BTPCommandBuilder {
	b.params = append(b.params, "--idp", idProvider)
	return b
}

// Build the final command string
func (b *BTPCommandBuilder) Build() ([]string, error) {
	if b.action == "" {
		return nil, errors.New("action is required")
	}

	cmdList := []string{"btp"}
	cmdList = append(cmdList, b.options...)
	cmdList = append(cmdList, b.action)
	if b.target != "" {
		cmdList = append(cmdList, b.target)
	}
	cmdList = append(cmdList, b.params...)

	return cmdList, nil
}

type BTPCommandBuilder struct {
	options []string
	action  string
	target  string
	params  []string
}
