package mongoproto

import "fmt"

var ErrNotMsg = fmt.Errorf("buffer is too small to be a Mongo message")

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
		return nil, ErrNotMsg
	}
	var m MsgHeader
	m.FromWire(b)

	var result Op
	switch m.OpCode {
	case OpCodeQuery:
		result = &OpQuery{Header: m}
	case OpCodeReply:
		result = &OpReply{Header: m}
	default:
		return nil, ErrUnknownOpcode(m.OpCode)
	}
	result.FromWire(b[MsgHeaderLen:])
	return result, nil
}
