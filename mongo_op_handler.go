package mongocaputils

import (
	"io"
	"log"

	"code.google.com/p/gopacket"
	"code.google.com/p/gopacket/tcpassembly"
	"code.google.com/p/gopacket/tcpassembly/tcpreader"

	"github.com/tmc/mongocaputils/mongoproto"
)

// TODO(tmc): reorder ops according to frame timings

type mongoOpStream struct {
	Ops chan mongoproto.Op
}

func NewMongoOpStream() *mongoOpStream {
	return &mongoOpStream{make(chan mongoproto.Op)}
}

func (s *mongoOpStream) New(a, b gopacket.Flow) tcpassembly.Stream {
	r := tcpreader.NewReaderStream()
	go s.handleStream(&r)
	return &r
}

func (s *mongoOpStream) Close() error {
	close(s.Ops)
	return nil
}

func (s *mongoOpStream) handleStream(r io.Reader) {
	for {
		op, err := mongoproto.OpFromReader(r)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("Error parsing op:", err)
			return
		}
		s.Ops <- op
	}
}
