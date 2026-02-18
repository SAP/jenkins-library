package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

// AvailableFlagValues returns all flags incl. values which are available to the command.
func AvailableFlagValues(cmd *cobra.Command, filters *StepFilters) map[string]any {
	flagValues := map[string]any{}
	flags := cmd.Flags()
	//only check flags where value has been set
	flags.Visit(func(pflag *flag.Flag) {

		switch pflag.Value.Type() {
		case "string":
			flagValues[pflag.Name] = pflag.Value.String()
		case "stringSlice":
			flagValues[pflag.Name], _ = flags.GetStringSlice(pflag.Name)
		case "bool":
			flagValues[pflag.Name], _ = flags.GetBool(pflag.Name)
		case "int":
			flagValues[pflag.Name], _ = flags.GetInt(pflag.Name)
		default:
			fmt.Printf("Meta data type not set or not known: '%v'\n", pflag.Value.Type())
			os.Exit(1)
		}
		filters.Parameters = append(filters.Parameters, pflag.Name)
	})
	return flagValues
}

// MarkFlagsWithValue marks a flag as changed if value is available for the flag through the step configuration.
func MarkFlagsWithValue(cmd *cobra.Command, stepConfig StepConfig) {
	flags := cmd.Flags()
	flags.VisitAll(func(pflag *flag.Flag) {
		//mark as available in case default is available or config is available
		if len(pflag.Value.String()) > 0 || stepConfig.Config[pflag.Name] != nil {
			pflag.Changed = true
		}
	})
}
