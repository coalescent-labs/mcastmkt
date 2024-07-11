package any

import (
	"github.com/spf13/cobra"
)

var (
	AnyCmd = &cobra.Command{
		Use:   "any",
		Short: "Generic multicast commands without enter in specific market protocol and conversion",
		Long:  ``,
	}
)

func init() {
	// Add subcommands here
	AnyCmd.AddCommand(listenCmd)
	AnyCmd.AddCommand(sendCmd)

}
