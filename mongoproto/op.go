package mongoproto

import (
	"fmt"
)

type Op interface {
	OpCode() OpCode
	ToWire() []byte
	FromWire(b []byte)
}

type ErrUnknownOpcode int

func (e ErrUnknownOpcode) Error() string {
	return fmt.Sprintf("Unknown opcode %d", e)
}

// OpFromWire reads an Op from a byte slice
func OpFromWire(b []byte) (Op, error) {
	if len(b) < MsgHeaderLen {
		return nil, fmt.Errorf("buffer too small, need at least %d, got %d",
			MsgHeaderLen, len(b))
	}
	var m MsgHeader
	m.MessageLength = getInt32(b, 0)
	m.RequestID = getInt32(b, 4)
	m.ResponseTo = getInt32(b, 8)
	m.OpCode = OpCode(getInt32(b, 12))

	switch m.OpCode {
	default:
		return nil, ErrUnknownOpcode(m.OpCode)
	}
}
