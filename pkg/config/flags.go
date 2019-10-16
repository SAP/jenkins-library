package config

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

// AvailableFlagValues returns all flags incl. values which are available to the command.
func AvailableFlagValues(cmd *cobra.Command, filters *StepFilters) map[string]interface{} {
	flagValues := map[string]interface{}{}
	flags := cmd.Flags()
	//only check flags where value has been set
	flags.Visit(func(pflag *flag.Flag) {
		flagValues[pflag.Name] = pflag.Value
		filters.Parameters = append(filters.Parameters, pflag.Name)
	})
	return flagValues
}

// MarkFlagsWithValue marks a flag as changed if value is available for the flag through the step configuration.
func MarkFlagsWithValue(cmd *cobra.Command, stepConfig StepConfig) {
	flags := cmd.Flags()
	flags.VisitAll(func(pflag *flag.Flag) {
		if stepConfig.Config[pflag.Name] != nil {
			pflag.Changed = true
		}
	})
}
