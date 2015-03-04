package mongocaputils

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"container/heap"

	"code.google.com/p/gopacket"
	"code.google.com/p/gopacket/tcpassembly"

	"github.com/tmc/mongocaputils/mongoproto"
	"github.com/tmc/mongocaputils/tcpreaderwrapper"
)

// TODO(tmc): reorder ops according to frame timings

type mongoOpStream struct {
	Ops chan mongoproto.Op

	firstSeen    time.Time
	opsWithTimes chan OpWithTime
	opHeap       *orderedOps

	started bool
	mu      sync.Mutex // for debugging
}

func NewMongoOpStream(heapBufSize int) *mongoOpStream {
	h := make(orderedOps, 0, heapBufSize)
	s := &mongoOpStream{
		Ops:          make(chan mongoproto.Op),
		opsWithTimes: make(chan OpWithTime),
		opHeap:       &h,
	}
	heap.Init(s.opHeap)
	go s.handleOps()
	return s
}

func (s *mongoOpStream) New(a, b gopacket.Flow) tcpassembly.Stream {
	r := tcpreaderwrapper.NewReaderStreamWrapper()
	if !s.started {
		s.started = true
	}
	log.Println("starting stream", a, b)
	go s.handleStream(&r)
	return &r
}

func (s *mongoOpStream) Close() error {
	close(s.opsWithTimes)
	s.opsWithTimes = nil
	return nil
}

func (s *mongoOpStream) SetFirstSeen(t time.Time) {
	s.firstSeen = t
}

func (s *mongoOpStream) handleOps() {
	defer close(s.Ops)
	for op := range s.opsWithTimes {
		heap.Push(s.opHeap, op)
		if len(*s.opHeap) >= cap(*s.opHeap) {
			op := heap.Pop(s.opHeap).(OpWithTime)
			fmt.Printf("%f %v\n", float64(op.Seen.Sub(s.firstSeen))/10e8, op.Op)
			s.Ops <- op.Op
			//s.Ops <- s.opHeap.Pop().(OpWithTime).Op
		}
	}
	for len(*s.opHeap) > 0 {
		op := heap.Pop(s.opHeap).(OpWithTime)
		fmt.Printf("%f %v \n", float64(op.Seen.Sub(s.firstSeen))/10e8, op.Op)
		s.Ops <- op.Op
		//s.Ops <- s.opHeap.Pop().(OpWithTime).Op
	}
}

func (s *mongoOpStream) readOp(r io.Reader) (mongoproto.Op, error) {
	return mongoproto.OpFromReader(r)
}

func (s *mongoOpStream) handleStream(r *tcpreaderwrapper.ReaderStreamWrapper) {
	for {
		op, err := s.readOp(r)
		if err == io.EOF {
			log.Println("stopping")
			discarded, err := ioutil.ReadAll(r)
			fmt.Println("discarded ", len(discarded), err)
			return
		}
		if err != nil {
			log.Println("error parsing op:", err)
			return
		}
		seen := time.Now()
		for _, r := range r.Reassemblies {
			if r.NumBytes > 0 {
				seen = r.Seen
			}
		}
		s.opsWithTimes <- OpWithTime{op, seen}
		r.Reassemblies = r.Reassemblies[:0]
	}
}
