package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// CommandLineCompletionCommand allows to generate convenience scripts for using the piper cli in a shell.
// See https://github.com/spf13/cobra/blob/master/shell_completions.md for docs on the subject.
func CommandLineCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

$ source <(piper completion bash)

# To load completions for each session, execute once:
Linux:
  $ piper completion bash > /etc/bash_completion.d/piper
MacOS:
  $ piper completion bash > /usr/local/etc/bash_completion.d/piper

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ piper completion zsh > "${fpath[1]}/_piper"

# You will need to start a new shell for this setup to take effect.

Fish:

$ piper completion fish | source

# To load completions for each session, execute once:
$ piper completion fish > ~/.config/fish/completions/piper.fish
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				_ = cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				_ = cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				_ = cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				_ = cmd.Root().GenPowerShellCompletion(os.Stdout)
			}
		},
	}
}
