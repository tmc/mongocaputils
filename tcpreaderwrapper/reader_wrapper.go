// package tcpreaderwrapper wraps a gopacket tcpassembly.tcpreader.ReaderStream
package tcpreaderwrapper

import (
	"code.google.com/p/gopacket/tcpassembly"
	"code.google.com/p/gopacket/tcpassembly/tcpreader"
)

type ReaderStreamWrapper struct {
	tcpreader.ReaderStream
	Reassemblies []ReassemblyInfo
}

// NewReaderStream returns a new ReaderStreamWrapper object.
func NewReaderStreamWrapper() ReaderStreamWrapper {
	r := ReaderStreamWrapper{
		ReaderStream: tcpreader.NewReaderStream(),
		Reassemblies: make([]ReassemblyInfo, 0),
	}
	return r
}

// Reassembled implements tcpassembly.Stream's Reassembled function.
func (r *ReaderStreamWrapper) Reassembled(reassembly []tcpassembly.Reassembly) {
	// keep track of sizes and times to reconstruct
	for _, re := range reassembly {
		r.Reassemblies = append(r.Reassemblies, newReassemblyInfo(re))
	}
	r.ReaderStream.Reassembled(reassembly)
}
