package any

import (
	"github.com/coalescent-labs/mcastmkt/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

var (
	listenAddress           string
	listenInterface         string
	listenDumpBytes         bool
	listenReceiveBufferSize int

	listenStatsInterval uint64 = 30

	listenTotalNumBytes   uint64 = 0
	listenNumBytes        uint64 = 0
	listenNumPackets      uint64 = 0
	listenTotalNumPackets uint64 = 0

	listenCmd = &cobra.Command{
		Use:   "listen",
		Short: "Listen multicast stream and dump statistics and data",
		Long:  ``,
		RunE:  listen,
	}
)

const (
	listenMaxDatagramSize = 1024 * 8
)

func listenStatsPrinter() {
	for range time.Tick(time.Second * time.Duration(listenStatsInterval)) {
		recvMsg := atomic.SwapUint64(&listenNumPackets, 0)
		recvTotalMsg := atomic.SwapUint64(&listenTotalNumPackets, 0)
		recvBytes := atomic.SwapUint64(&listenNumBytes, 0)
		recvTotalBytes := atomic.SwapUint64(&listenTotalNumBytes, 0)
		log.Printf("STAT Recv msg: %d [Tot %d], Recv bytes: %s [Tot: %s]",
			recvMsg, recvTotalMsg, util.ByteCountIEC(recvBytes), util.ByteCountIEC(recvTotalBytes))
	}
}

func listen(*cobra.Command, []string) error {
	// Parse the string address
	addr, err := net.ResolveUDPAddr("udp4", listenAddress)
	if err != nil {
		return err
	}

	var intf *net.Interface = nil

	if listenInterface != "" {
		intf, err = util.GetInterfaceFromIPorName(listenInterface)
		if err != nil {
			return err
		}
	}

	conn, err := net.ListenPacket("udp4", listenAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	if listenReceiveBufferSize > 0 {
		if err := util.SetReceiveBuffer(conn, listenReceiveBufferSize); err != nil {
			return err
		}
	}

	packetConn := ipv4.NewPacketConn(conn)
	if err := packetConn.JoinGroup(intf, addr); err != nil {
		return err
	}
	defer packetConn.LeaveGroup(intf, addr)

	err = packetConn.SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst|ipv4.FlagInterface, true)
	if err != nil {
		return err
	}
	buffer := make([]byte, listenMaxDatagramSize)

	go listenStatsPrinter()

	log.Printf("Listening to %s@%s  %v\n", listenAddress, util.StringIfEmpty(listenInterface, "default"), intf)

	// Loop forever reading from the socket
	for {
		numBytes, cm, srcAddr, err := packetConn.ReadFrom(buffer)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		atomic.AddUint64(&listenTotalNumPackets, 1)
		atomic.AddUint64(&listenTotalNumBytes, uint64(numBytes))

		if !cm.Dst.IsMulticast() {
			continue
		}
		if !cm.Dst.Equal(addr.IP) {
			// unknown group, discard
			continue
		}

		atomic.AddUint64(&listenNumPackets, 1)
		atomic.AddUint64(&listenNumBytes, uint64(numBytes))

		if listenDumpBytes {
			log.Printf(strings.Repeat("-", 80))
			log.Printf("addr: %v, numBytes: %d\n", srcAddr, numBytes)
			util.DumpByteSlice(buffer[:numBytes])
		}
	}
}

func init() {
	listenCmd.PersistentFlags().StringVarP(&listenAddress, "address", "a", "224.0.50.59:59001", "The multicast address and port")
	listenCmd.PersistentFlags().StringVarP(&listenInterface, "interface", "i", "", "The multicast listener interface name or IP address")
	listenCmd.PersistentFlags().BoolVarP(&listenDumpBytes, "dump", "d", false, "Dump the raw bytes of the message")
	listenCmd.PersistentFlags().IntVarP(&listenReceiveBufferSize, "receive-buffer-size", "r", 0, "Socket receive buffer size in bytes (0 use system default)")
	listenCmd.PersistentFlags().Uint64VarP(&listenStatsInterval, "stats-interval", "s", 30, "Statistics print interval in seconds")
	_ = listenCmd.MarkPersistentFlagRequired("address")
	_ = viper.BindPFlag("address", listenCmd.PersistentFlags().Lookup("address"))
	_ = viper.BindPFlag("interface", listenCmd.PersistentFlags().Lookup("interface"))
	_ = viper.BindPFlag("dump", listenCmd.PersistentFlags().Lookup("dump"))
	_ = viper.BindPFlag("receive-buffer-size", listenCmd.PersistentFlags().Lookup("receive-buffer-size"))
	_ = viper.BindPFlag("stats-interval", listenCmd.PersistentFlags().Lookup("stats-interval"))
}
