package any

import (
	"fmt"
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
	sendAddress   string
	sendInterface string
	sendDumpBytes bool

	sendInterval      uint64 = 1000
	sendStatsInterval uint64 = 30
	sendTtl           int    = 1
	sendText          string = "This is test number: {c}"

	sendNumBytes   uint64 = 0
	sendNumPackets uint64 = 0

	sendCmd = &cobra.Command{
		Use:   "send",
		Short: "Send multicast test message continuously in a loop at specified interval until the program is terminated",
		Long:  ``,
		RunE:  send,
	}
)

func sendStatsPrinter() {
	for range time.Tick(time.Second * time.Duration(sendStatsInterval)) {
		sentMsg := atomic.SwapUint64(&sendNumPackets, 0)
		sentBytes := atomic.SwapUint64(&sendNumBytes, 0)
		log.Printf("STAT Send msg: %d, Send bytes: %s",
			sentMsg, util.ByteCountIEC(sentBytes))
	}
}

func send(*cobra.Command, []string) error {
	// Parse the string address
	addr, err := net.ResolveUDPAddr("udp4", sendAddress)
	if err != nil {
		return err
	}

	var intf *net.Interface = nil

	if sendInterface != "" {
		intf, err = util.GetInterfaceFromIPorName(sendInterface)
		if err != nil {
			return err
		}
	}

	// create a UDP connection
	conn, err := net.ListenPacket("udp4", "")
	if err != nil {
		return err
	}

	defer conn.Close()

	packetConn := ipv4.NewPacketConn(conn)
	if intf != nil {
		err = packetConn.SetMulticastInterface(intf)
		if err != nil {
			return err
		}
	}

	err = packetConn.SetMulticastTTL(sendTtl)
	if err != nil {
		return err
	}

	err = packetConn.SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst|ipv4.FlagInterface, true)
	if err != nil {
		return err
	}

	go sendStatsPrinter()

	log.Printf("Sending to %s@%s  %v\n", sendAddress, util.StringIfEmpty(sendInterface, "default"), intf)

	var text func(int) string
	if strings.Contains(sendText, "{c}") {
		subStr := strings.Replace(sendText, "{c}", "%d", 1)
		text = func(x int) string {
			return fmt.Sprintf(subStr, x)
		}
	} else {
		text = func(x int) string { return sendText }
	}

	var c = 0
	var numBytes int

	// loop forever sending messages
	for range time.Tick(time.Millisecond * time.Duration(sendInterval)) {
		c++
		msg := []byte(text(c))
		numBytes, err = packetConn.WriteTo(msg, nil, addr)
		if err != nil {
			log.Fatal("Write failed:", err)
		}

		atomic.AddUint64(&sendNumPackets, 1)
		atomic.AddUint64(&sendNumBytes, uint64(numBytes))

		if sendDumpBytes {
			log.Printf(strings.Repeat("-", 80))
			util.DumpByteSlice(msg)
		}
	}

	return nil
}

func init() {
	sendCmd.PersistentFlags().StringVarP(&sendAddress, "address", "a", "224.0.50.59:59001", "The multicast address and port")
	sendCmd.PersistentFlags().StringVarP(&sendInterface, "interface", "i", "", "The multicast send interface name or IP address")
	sendCmd.PersistentFlags().BoolVarP(&sendDumpBytes, "dump", "d", false, "Dump the raw bytes of the sent message")
	sendCmd.PersistentFlags().Uint64VarP(&sendInterval, "interval", "n", 1000, "Interval in milliseconds between sending messages")
	sendCmd.PersistentFlags().IntVarP(&sendTtl, "ttl", "t", 1, "Time to live")
	sendCmd.PersistentFlags().StringVar(&sendText, "text", "This is test number: {c}", "Text/data to send to the receiver. Use '{c}' to send counter")
	sendCmd.PersistentFlags().Uint64VarP(&sendStatsInterval, "stats-interval", "s", 30, "Statistics print interval in seconds")
	_ = sendCmd.MarkPersistentFlagRequired("address")
	_ = viper.BindPFlag("address", listenCmd.PersistentFlags().Lookup("address"))
	_ = viper.BindPFlag("interface", listenCmd.PersistentFlags().Lookup("interface"))
	_ = viper.BindPFlag("dump", listenCmd.PersistentFlags().Lookup("dump"))
	_ = viper.BindPFlag("interval", listenCmd.PersistentFlags().Lookup("interval"))
	_ = viper.BindPFlag("ttl", listenCmd.PersistentFlags().Lookup("ttl"))
	_ = viper.BindPFlag("text", listenCmd.PersistentFlags().Lookup("text"))
	_ = viper.BindPFlag("stats-interval", listenCmd.PersistentFlags().Lookup("stats-interval"))
}
