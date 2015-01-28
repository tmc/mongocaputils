package main

import (
	"flag"
	"fmt"
	"os"

	"code.google.com/p/gopacket/pcap"
	"github.com/tmc/mongocaputils"
)

var (
	pcapFile = flag.String("f", "-", "pcap file (or '-' for stdin)")
)

func main() {
	flag.Parse()

	pcap, err := pcap.OpenOffline(*pcapFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error opening pcap file:", err)
		os.Exit(1)
	}
	h := mongocaputils.NewPacketHandler(pcap)
	m := mongocaputils.NewMongoOpStream()

	ch := make(chan struct{})
	go func() {
		defer close(ch)
		i := 0
		for op := range m.Ops {
			i++
			fmt.Println(i, op)
		}
	}()

	if err := h.Handle(m, -1); err != nil {
		fmt.Fprintln(os.Stderr, "mongocapcat: error handling packet stream:", err)
	}
	<-ch
}
