package mongoproto

import "gopkg.in/mgo.v2/bson"

const (
	OpReplyCursorNotFound   OpReplyFlags = 1 << iota // Set when getMore is called but the cursor id is not valid at the server. Returned with zero results.
	OpReplyQueryFailure                             // Set when query failed. Results consist of one document containing an “$err” field describing the failure.
	OpReplyShardConfigStale                         //Drivers should ignore this. Only mongos will ever see this set, in which case, it needs to update config from the server.
	OpReplyAwaitCapable                             //Set when the server supports the AwaitData Query option. If it doesn’t, a client should sleep a little between getMore’s of a Tailable cursor. Mongod version 1.6 supports AwaitData and thus always sets AwaitCapable.
)

type OpReplyFlags int32

// OpReply is sent by the database in response to an OpQuery or OpGetMore message.
// http://docs.mongodb.org/meta-driver/latest/legacy/mongodb-wire-protocol/#op-reply
type OpReply struct {
	Header         MsgHeader
	Message        string
	Flags          OpReplyFlags
	CursorID       int64     // cursor id if client needs to do get more's
	StartingFrom   int32     // where in the cursor this reply is starting
	NumberReturned int32     // number of documents in the reply
	Documents      []*bson.D // documents
}
