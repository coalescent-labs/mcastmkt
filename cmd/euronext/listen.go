package euronext

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	listenAddress           string
	listenInterface         string
	listenDumpBytes         bool
	listenReceiveBufferSize int
	listenStatsInterval     uint64 = 30

	listenCmd = &cobra.Command{
		Use:   "listen",
		Short: "Listen Euronext Optiq multicast stream and dump out of sequence messages",
		Long:  ``,
	}
)

func init() {
	listenCmd.PersistentFlags().StringVarP(&listenAddress, "address", "a", "224.0.50.59:59001", "The multicast address and port")
	listenCmd.PersistentFlags().StringVarP(&listenInterface, "interface", "i", "", "The multicast listener interface name or IP address")
	listenCmd.PersistentFlags().BoolVarP(&listenDumpBytes, "dump", "d", false, "Dump the raw bytes of the message")
	listenCmd.PersistentFlags().IntVarP(&listenReceiveBufferSize, "receive-buffer-size", "r", 0, "Socket receive buffer size in bytes (0 use system default)")
	listenCmd.PersistentFlags().Uint64VarP(&listenStatsInterval, "stats-interval", "s", 30, "Interval between printing stats (seconds). Default is 30s")
	_ = listenCmd.MarkPersistentFlagRequired("address")
	_ = viper.BindPFlag("address", listenCmd.PersistentFlags().Lookup("address"))
	_ = viper.BindPFlag("interface", listenCmd.PersistentFlags().Lookup("interface"))
	_ = viper.BindPFlag("dump", listenCmd.PersistentFlags().Lookup("dump"))
	_ = viper.BindPFlag("receive-buffer-size", listenCmd.PersistentFlags().Lookup("receive-buffer-size"))
	_ = viper.BindPFlag("stats-interval", listenCmd.PersistentFlags().Lookup("stats-interval"))

	// Add subcommands here
	listenCmd.AddCommand(listenMdgCmd)
}
