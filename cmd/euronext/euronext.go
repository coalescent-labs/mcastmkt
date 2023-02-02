package euronext

import (
	"github.com/spf13/cobra"
)

var (
	EuronextCmd = &cobra.Command{
		Use:   "euronext",
		Short: "Euronext optiq multicast commands",
		Long:  ``,
	}
)

func init() {
	// Add subcommands here
	EuronextCmd.AddCommand(listenCmd)

}
