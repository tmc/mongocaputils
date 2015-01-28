package mongoproto

import (
	"bytes"
	"io"
	"log"
)

const (
	OpInsertContinueOnError OpInsertFlags = 1 << iota
)

type OpInsertFlags int32

// OpInsert is used to insert one or more documents into a collection.
// http://docs.mongodb.org/meta-driver/latest/legacy/mongodb-wire-protocol/#op-insert
type OpInsert struct {
	Header             MsgHeader
	Flags              OpInsertFlags
	FullCollectionName string   // "dbname.collectionname"
	Documents          [][]byte // one or more documents to insert into the collection
}

func (op *OpInsert) OpCode() OpCode {
	return OpCodeInsert
}

func (op *OpInsert) FromReader(r io.Reader) error {
	var b [4]byte
	_, err := io.ReadFull(r, b[:])
	if err != nil {
		return err
	}
	op.Flags = OpInsertFlags(getInt32(b[:], 0))
	name, err := readCStringFromReader(r)
	if err != nil {
		return err
	}
	op.FullCollectionName = string(name)
	op.Documents = make([][]byte, 0)

	docLen := 0
	for len(name)+1+4+docLen < int(op.Header.MessageLength) {
		doc, err := ReadDocument(r)
		if err != nil {
			return err
		}
		docLen += len(doc)
		op.Documents = append(op.Documents, doc)
	}
	return nil
}

func (op *OpInsert) fromWire(b []byte) {
	if len(b) < 5 {
		return
	}
	op.Flags = OpInsertFlags(getInt32(b, 0))
	op.FullCollectionName = readCString(b[4:])
	b = b[len(op.FullCollectionName)+1:]
	op.Documents = make([][]byte, 0)
	offset := 0
	for len(b) > 0 {
		doc, err := ReadDocument(bytes.NewReader(b[offset:]))
		if err != nil {
			log.Println("doc err:", err, len(b[offset:]))
			break
		}
		offset += len(doc)
		op.Documents = append(op.Documents, doc)
	}
}

func (op *OpInsert) toWire() []byte {
	return nil
}
