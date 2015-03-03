package mongocaputils

import (
	"io"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"

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

type SetFirstSeener interface {
	SetFirstSeen(t time.Time)
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
		log.Println("flushing assembler.")
		log.Println("num flushed/closed:", assembler.FlushAll())
		log.Println("closing stream handler.")
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
	ticker := time.Tick(time.Second * 10)
	for {
		select {
		case pkt := <-source.Packets():
			if pkt == nil { // end of pcap file
				return nil
			}
			if tcpLayer := pkt.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				assembler.AssembleWithTimestamp(
					pkt.TransportLayer().TransportFlow(),
					tcpLayer.(*layers.TCP),
					pkt.Metadata().Timestamp)
			} else {
				spew.Dump("NONTCP:", pkt.Layers())
			}

			if count == 0 {
				if firstSeener, ok := streamHandler.(SetFirstSeener); ok {
					firstSeener.SetFirstSeen(pkt.Metadata().Timestamp)
				}
			}

			count++
			if numToHandle > 0 && count >= int64(numToHandle) {
				log.Println("Count exceeds requested packets, returning.")
				break
			}
		case <-ticker:
			log.Println("flushing old streams")
			assembler.FlushOlderThan(time.Now().Add(time.Second * -5))
		}
	}
	return nil
}
