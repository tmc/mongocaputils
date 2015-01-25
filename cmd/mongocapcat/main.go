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
	h := mongocaputils.NewPacketHandler(pcap, 100)
	m := mongocaputils.NewMongoOpHandler(h.Packets)
	go m.Loop()

	if err := h.Handle(0); err != nil {
		fmt.Fprintln(os.Stderr, "error handling packets:", err)
		os.Exit(1)
	}
	<-m.Finished
}
