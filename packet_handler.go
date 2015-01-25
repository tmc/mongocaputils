package mongocaputils

import (
	"io"
	"log"
	"time"

	"code.google.com/p/gopacket"
	"code.google.com/p/gopacket/pcap"
)

type PacketHandler struct {
	pcap       *pcap.Handle
	Packets    chan gopacket.Packet
	numDropped int64
}

func NewPacketHandler(pcapHandle *pcap.Handle, packetBufferSize int) *PacketHandler {
	return &PacketHandler{
		pcap:    pcapHandle,
		Packets: make(chan gopacket.Packet, packetBufferSize),
	}
}
func (m *PacketHandler) Handle(numToHandle int) error {
	count := int64(0)
	start := time.Now()
	defer func() {
		log.Println("Dropped", m.numDropped,
			"packets out of", count)
		runTime := float64(time.Now().Sub(start)) / float64(time.Second)
		log.Println("Processed",
			float64(count-m.numDropped)/runTime,
			"packets per second")
	}()
	if numToHandle > 0 {
		log.Println("Processing", numToHandle, "packets")
	}

	defer close(m.Packets)

	source := gopacket.NewPacketSource(m.pcap, m.pcap.LinkType())
	for {
		pkt, err := source.NextPacket()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println("Error:", err)
			continue
		}
		if numToHandle > 0 && count >= int64(numToHandle) {
			log.Println("Count exceeds requested packets, returning.")
			return nil
		}
		select {
		case m.Packets <- pkt:
		default:
			m.numDropped++
		}
		count++
	}
	return nil
}
