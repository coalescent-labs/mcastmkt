package eurex

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
	emdiTotalNumBytes   uint64 = 0
	emdiNumBytes        uint64 = 0
	emdiNumPackets      uint64 = 0
	emdiTotalNumPackets uint64 = 0
	emdiNumPacketsOoO   uint64 = 0
	emdiNumPacketsMessy uint64 = 0
	lastSeqNum          uint32 = 0

	listenEmdiCmd = &cobra.Command{
		Use:   "emdi",
		Short: "Listen Eurex EMDI multicast stream and dump out of sequence messages",
		Long: `Detecting duplicates and gaps by packet header.
Packets with the same SenderCompID (field length: 1 Byte) have contiguous sequence numbers per multicast address / port combination.`,
		RunE: listenEmdi,
	}
)

func statsPrinter() {
	for range time.Tick(time.Second * 60) {
		recvMsg := atomic.SwapUint64(&emdiNumPackets, 0)
		recvTotalMsg := atomic.SwapUint64(&emdiTotalNumPackets, 0)
		recvBytes := atomic.SwapUint64(&emdiNumBytes, 0)
		recvTotalBytes := atomic.SwapUint64(&emdiTotalNumBytes, 0)
		recvOoO := atomic.SwapUint64(&emdiNumPacketsOoO, 0)
		recvMessy := atomic.SwapUint64(&emdiNumPacketsMessy, 0)
		log.Printf("STAT Recv msg: %d [Tot %d], Recv bytes: %s [Tot: %s], Last seqNo: %d, OoO: %d, Messy: %d\n",
			recvMsg, recvTotalMsg, util.ByteCountIEC(recvBytes), util.ByteCountIEC(recvTotalBytes),
			lastSeqNum, recvOoO, recvMessy)
	}
}

func listenEmdi(*cobra.Command, []string) error {
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
		atomic.AddUint64(&emdiTotalNumPackets, 1)
		atomic.AddUint64(&emdiTotalNumBytes, uint64(numBytes))

		if !cm.Dst.IsMulticast() {
			continue
		}
		if !cm.Dst.Equal(addr.IP) {
			// unknown group, discard
			continue
		}

		atomic.AddUint64(&emdiNumPackets, 1)
		atomic.AddUint64(&emdiNumBytes, uint64(numBytes))
		if numBytes > 9 {
			//pmap := buffer[0]
			//tid := buffer[1]
			partitionId := buffer[2]
			senderCompId := buffer[3]
			length := buffer[4]
			seqNum := binary.BigEndian.Uint32(buffer[5:9])

			_, ok := cache.Get(seqNum)
			if ok {
				log.Printf("Duplicate message: %d\n", seqNum)
				continue
			}
			cache.Add(seqNum, true)

			if listenDumpBytes {
				log.Printf(strings.Repeat("-", 80))
				log.Printf("addr: %v, numBytes: %d, partitionId: %d, senderCompId: %d, length: %d, lastSeqNum: %d, seqNum: %d\n", srcAddr, numBytes, partitionId, senderCompId, length, lastSeqNum, seqNum)
				util.DumpByteSlice(buffer[:numBytes])
			}

			if lastSeqNum != 0 && seqNum > lastSeqNum && seqNum != lastSeqNum+1 {
				ooo := uint64(seqNum - lastSeqNum - 1)
				atomic.AddUint64(&emdiNumPacketsOoO, ooo)
				log.Printf("Out of sequence message: %d -> %d [%d]\n", lastSeqNum, seqNum, ooo)
			}
			if lastSeqNum != 0 && seqNum < lastSeqNum {
				atomic.AddUint64(&emdiNumPacketsMessy, 1)
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
