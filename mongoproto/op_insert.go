package mongoproto

import "gopkg.in/mgo.v2/bson"

const (
	OpInsertContinueOnError OpInsertFlags = 1 << iota
)

type OpInsertFlags int32

// OpInsert is used to insert one or more documents into a collection.
// http://docs.mongodb.org/meta-driver/latest/legacy/mongodb-wire-protocol/#op-insert
type OpInsert struct {
	Header             MsgHeader
	Flags              OpInsertFlags
	FullCollectionName string    // "dbname.collectionname"
	Documents          []*bson.D // one or more documents to insert into the collection
}
