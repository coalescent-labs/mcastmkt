package eurex

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	listenAddress           string
	listenInterface         string
	listenDumpBytes         bool
	listenReceiveBufferSize int

	listenCmd = &cobra.Command{
		Use:   "listen",
		Short: "Listen Eurex multicast stream and dump out of sequence messages",
		Long:  ``,
	}
)

func init() {
	listenCmd.PersistentFlags().StringVarP(&listenAddress, "address", "a", "224.0.50.59:59001", "The multicast address and port")
	listenCmd.PersistentFlags().StringVarP(&listenInterface, "interface", "i", "", "The multicast listener interface name or IP address")
	listenCmd.PersistentFlags().BoolVarP(&listenDumpBytes, "dump", "d", false, "Dump the raw bytes of the message")
	listenCmd.PersistentFlags().IntVarP(&listenReceiveBufferSize, "receiveBufferSize", "r", 0, "Socket receive buffer size in bytes (0 use system default)")
	_ = listenCmd.MarkPersistentFlagRequired("address")
	_ = viper.BindPFlag("address", listenCmd.PersistentFlags().Lookup("address"))
	_ = viper.BindPFlag("interface", listenCmd.PersistentFlags().Lookup("interface"))
	_ = viper.BindPFlag("dump", listenCmd.PersistentFlags().Lookup("dump"))

	// Add subcommands here
	listenCmd.AddCommand(listenEmdiCmd)

}
