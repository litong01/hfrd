package execute

import "github.com/spf13/cobra"

var executeCmd = &cobra.Command{
	Use:              "execute",
	Short:            "execute bash scripts or binary file , command ",
	TraverseChildren: true,
}

var (
	commandName   string
	commandParams string
)

const (
	COMMAND_NAME   = "commandName"
	COMMAND_PARAMS = "commandParams"
)

func Cmd() *cobra.Command {
	executeCmd.AddCommand(cmCmd())
	return executeCmd
}
