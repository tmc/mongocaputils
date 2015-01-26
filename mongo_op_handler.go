package mongocaputils

import (
	"errors"
	"fmt"
	"log"

	"code.google.com/p/gopacket"

	"github.com/tmc/mongocaputils/mongoproto"
)

var ErrNoPayload = errors.New("mongocaputils: packet has no payload")

type MongoOpHandler struct {
	packets  chan gopacket.Packet
	Finished chan struct{}
}

func NewMongoOpHandler(packets chan gopacket.Packet) *MongoOpHandler {
	return &MongoOpHandler{
		packets:  packets,
		Finished: make(chan struct{}),
	}
}

func (m *MongoOpHandler) Loop() {
	defer close(m.Finished)
	count := 0
	for {
		p, ok := <-m.packets
		if !ok {
			return
		}
		op, err := m.HandlePacket(p)
		if err == ErrNoPayload {
			continue
		}
		if err == mongoproto.ErrNotMsg {
			continue
		}
		if err != nil {
			log.Println("error handling mongo packet:", err)
			continue
		}
		count++
		fmt.Printf("%3d: %v\n", count, op)
	}
}

func (m *MongoOpHandler) HandlePacket(p gopacket.Packet) (mongoproto.Op, error) {
	// assume tcp
	if appLayer := p.ApplicationLayer(); appLayer != nil {
		return mongoproto.OpFromWire(p.ApplicationLayer().Payload())
	}
	if errLayer := p.ErrorLayer(); errLayer != nil {
		return nil, fmt.Errorf("error parsing packet:", errLayer)
	}
	return nil, ErrNoPayload
}
