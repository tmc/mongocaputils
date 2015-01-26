package mongoproto

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/mongodb/mongo-tools/common/bsonutil"
	"gopkg.in/mgo.v2/bson"
)

const (
	_ OpQueryFlags = 1 << iota

	OpQueryTailableCursor  // Tailable means cursor is not closed when the last data is retrieved. Rather, the cursor marks the final object’s position. You can resume using the cursor later, from where it was located, if more data were received. Like any “latent cursor”, the cursor may become invalid at some point (CursorNotFound) – for example if the final object it references were deleted.
	OpQuerySlaveOk         // Allow query of replica slave. Normally these return an error except for namespace “local”.
	OpQueryOplogReplay     // Internal replication use only - driver should not set
	OpQueryNoCursorTimeout // The server normally times out idle cursors after an inactivity period (10 minutes) to prevent excess memory use. Set this option to prevent that.
	OpQueryAwaitData       // Use with TailableCursor. If we are at the end of the data, block for a while rather than returning no data. After a timeout period, we do return as normal.
	OpQueryExhaust         // Stream the data down full blast in multiple “more” packages, on the assumption that the client will fully read all data queried. Faster when you are pulling a lot of data and know you want to pull it all down. Note: the client is not allowed to not read all the data unless it closes the connection.
	OpQueryPartial         // Get partial results from a mongos if some shards are down (instead of throwing an error)
)

type OpQueryFlags int32

// OpQuery is used to query the database for documents in a collection.
// http://docs.mongodb.org/meta-driver/latest/legacy/mongodb-wire-protocol/#op-query
type OpQuery struct {
	Header               MsgHeader
	Flags                OpQueryFlags
	FullCollectionName   string // "dbname.collectionname"
	NumberToSkip         int32  // number of documents to skip
	NumberToReturn       int32  // number of documents to return
	Query                []byte // query object
	ReturnFieldsSelector []byte // Optional. Selector indicating the fields to return
}

func (op *OpQuery) String() string {
	var query interface{}
	bson.Unmarshal(op.Query, &query)
	queryAsJSON, err := bsonutil.ConvertBSONValueToJSON(query)
	if err != nil {
		return fmt.Sprintf("%#v - %v", op, err)
	}
	asJSON, err := json.Marshal(queryAsJSON)
	if err != nil {
		return fmt.Sprintf("%#v - %v", op, err)
	}
	return fmt.Sprintf("OpQuery %v %v", op.FullCollectionName, string(asJSON))
}

func (op *OpQuery) OpCode() OpCode {
	return OpCodeQuery
}

func (op *OpQuery) FromWire(b []byte) {
	if len(b) < 16 {
		return
	}
	op.Flags = OpQueryFlags(getInt32(b, 0))
	op.FullCollectionName = readCString(b[4:])

	b = b[4+len(op.FullCollectionName)+1:]
	op.NumberToSkip = getInt32(b, 0)
	op.NumberToReturn = getInt32(b, 4)

	b = b[8:]
	op.Query, _ = ReadDocument(bytes.NewReader(b))
	b = b[len(op.Query):]

	if len(b) < 4 {
		return
	}
	returnFields, err := ReadDocument(bytes.NewReader(b))
	if err == nil {
		op.ReturnFieldsSelector = returnFields
	}
}

func (op *OpQuery) ToWire() []byte {
	return nil
}
