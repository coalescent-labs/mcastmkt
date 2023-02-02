package eurex

import (
	"github.com/spf13/cobra"
)

var (
	EurexCmd = &cobra.Command{
		Use:   "eurex",
		Short: "Eurex multicast commands",
		Long:  ``,
	}
)

func init() {
	// Add subcommands here
	EurexCmd.AddCommand(listenCmd)

}
