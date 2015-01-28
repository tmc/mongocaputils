package mongocaputils

import (
	"io"
	"log"
	"time"

	"code.google.com/p/gopacket"
	"code.google.com/p/gopacket/layers"
	"code.google.com/p/gopacket/pcap"
	"code.google.com/p/gopacket/tcpassembly"
)

type PacketHandler struct {
	pcap       *pcap.Handle
	numDropped int64
}

func NewPacketHandler(pcapHandle *pcap.Handle) *PacketHandler {
	return &PacketHandler{
		pcap: pcapHandle,
	}
}

type StreamHandler interface {
	tcpassembly.StreamFactory
	io.Closer
}

func (m *PacketHandler) Handle(streamHandler StreamHandler, numToHandle int) error {

	count := int64(0)
	start := time.Now()
	if numToHandle > 0 {
		log.Println("Processing", numToHandle, "packets")
	}

	source := gopacket.NewPacketSource(m.pcap, m.pcap.LinkType())
	streamPool := tcpassembly.NewStreamPool(streamHandler)
	assembler := tcpassembly.NewAssembler(streamPool)
	defer func() {
		assembler.FlushAll()
		streamHandler.Close()
	}()
	defer func() {
		log.Println("Dropped", m.numDropped,
			"packets out of", count)
		runTime := float64(time.Now().Sub(start)) / float64(time.Second)
		log.Println("Processed",
			float64(count-m.numDropped)/runTime,
			"packets per second")
	}()
	for {
		pkt, err := source.NextPacket()

		if err == io.EOF {
			break
		} else if err != nil {
			log.Println("Error:", err)
			continue
		}

		if tcpLayer := pkt.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			assembler.AssembleWithTimestamp(
				pkt.TransportLayer().TransportFlow(),
				tcpLayer.(*layers.TCP),
				pkt.Metadata().Timestamp)
		}

		count++
		if numToHandle > 0 && count >= int64(numToHandle) {
			log.Println("Count exceeds requested packets, returning.")
			break
		}
	}
	return nil
}
