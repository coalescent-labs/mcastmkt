package euronext

import (
	"encoding/binary"
	"github.com/coalescent-labs/mcastmkt/pkg/util"
	"github.com/hashicorp/golang-lru"
	"github.com/spf13/cobra"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

const (
	listenMaxDatagramSize = 1024 * 8
)

var (
	mdgTotalNumBytes   uint64 = 0
	mdgNumBytes        uint64 = 0
	mdgNumPackets      uint64 = 0
	mdgTotalNumPackets uint64 = 0
	mdgNumPacketsOoO   uint64 = 0
	mdgNumPacketsMessy uint64 = 0
	lastSeqNum         uint32 = 0

	listenMdgCmd = &cobra.Command{
		Use:   "mdg",
		Short: "Listen Euronext Optiq MDG multicast stream and dump out of sequence messages",
		Long:  `Detecting duplicates and gaps by packet header.`,
		RunE:  listenMdg,
	}
)

func statsPrinter() {
	for range time.Tick(time.Second * 60) {
		recvMsg := atomic.SwapUint64(&mdgNumPackets, 0)
		recvTotalMsg := atomic.SwapUint64(&mdgTotalNumPackets, 0)
		recvBytes := atomic.SwapUint64(&mdgNumBytes, 0)
		recvTotalBytes := atomic.SwapUint64(&mdgTotalNumBytes, 0)
		recvOoO := atomic.SwapUint64(&mdgNumPacketsOoO, 0)
		recvMessy := atomic.SwapUint64(&mdgNumPacketsMessy, 0)
		log.Printf("STAT Recv msg: %d [Tot %d], Recv bytes: %s [Tot: %s], Last seqNo: %d, OoO: %d, Messy: %d\n",
			recvMsg, recvTotalMsg, util.ByteCountIEC(recvBytes), util.ByteCountIEC(recvTotalBytes),
			lastSeqNum, recvOoO, recvMessy)
	}
}

func listenMdg(*cobra.Command, []string) error {
	// Parse the string address
	addr, err := net.ResolveUDPAddr("udp4", listenAddress)
	if err != nil {
		return err
	}

	var intf *net.Interface = nil

	if listenInterface != "" {
		intf, err = net.InterfaceByName(listenInterface)
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

	// lru cache for sequence numbers duplicates check
	cache, _ := lru.New(2048)

	go statsPrinter()

	log.Printf("Listening to %s@%s  %v\n", listenAddress, util.StringIfEmpty(listenInterface, "default"), intf)

	lastSeqNum = 0
	// Loop forever reading from the socket
	for {
		numBytes, cm, srcAddr, err := packetConn.ReadFrom(buffer)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		atomic.AddUint64(&mdgTotalNumPackets, 1)
		atomic.AddUint64(&mdgTotalNumBytes, uint64(numBytes))

		if !cm.Dst.IsMulticast() {
			continue
		}
		if !cm.Dst.Equal(addr.IP) {
			// unknown group, discard
			continue
		}

		atomic.AddUint64(&mdgNumPackets, 1)
		atomic.AddUint64(&mdgNumBytes, uint64(numBytes))
		if numBytes >= 16 {
			packetTime := binary.LittleEndian.Uint64(buffer[0:8])
			/**
			Used to flag information (Little-Endian):
			- Bit 0: Compression
			 - 0 = body of the packet is not compressed (the body is the packet without the packet header)
			 - 1 = body of the packet is compressed
			- Bit 1 to 3: will be set to 0 every morning and incremented for each restart of MDG in the same day (wrapping to 0 if the field overflows
			- Bit 4 to 6: used if the Packet Sequence Number (PSN) goes over (2^32)-1. They are PSN high weight bits
			- Bit 7: is set to 1 when in the packet there is a Start Of Snapshot (2101) message, 0 otherwise
			- Bit 8: is set to 1 when in the packet there is an End Of Snapshot (2102) message, 0 otherwise
			- Bit 9: is set to 1 when in the packet there is a Health Status (1103) message, Start Of Day (1101) message or End Of Day (1102) message, 0 otherwise
			- Bit 10 to 15: for future use
			*/
			packetFlags := binary.LittleEndian.Uint16(buffer[12:14])
			seqNum := binary.LittleEndian.Uint32(buffer[8:12])
			channelId := binary.LittleEndian.Uint16(buffer[14:16])

			_, ok := cache.Get(seqNum)
			if ok {
				log.Printf("Duplicate message: %d\n", seqNum)
				continue
			}
			cache.Add(seqNum, true)

			if listenDumpBytes {
				log.Printf(strings.Repeat("-", 80))
				log.Printf("addr: %v, numBytes: %d, time: %d, channelId: %d, flags: %x, lastSeqNum: %d, seqNum: %d\n", srcAddr, numBytes, packetTime, channelId, packetFlags, lastSeqNum, seqNum)
				util.DumpByteSlice(buffer[:numBytes])
			}

			if lastSeqNum != 0 && seqNum > lastSeqNum && seqNum != lastSeqNum+1 {
				ooo := uint64(seqNum - lastSeqNum - 1)
				atomic.AddUint64(&mdgNumPacketsOoO, ooo)
				log.Printf("Out of sequence message: %d -> %d [%d]\n", lastSeqNum, seqNum, ooo)
			}
			if lastSeqNum != 0 && seqNum < lastSeqNum {
				atomic.AddUint64(&mdgNumPacketsMessy, 1)
				log.Printf("Messy message: %d\n", seqNum)
			}
			if seqNum > lastSeqNum {
				lastSeqNum = seqNum
			}
		} else {
			log.Fatalf("ReadFromUDP failed wrong num bytes: %d", numBytes)
		}
	}
}
