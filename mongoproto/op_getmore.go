package mongoproto

// OpGetMore is used to query the database for documents in a collection.
// http://docs.mongodb.org/meta-driver/latest/legacy/mongodb-wire-protocol/#op-get-more
type OpGetMore struct {
	Header             MsgHeader
	FullCollectionName string // "dbname.collectionname"
	NumberToReturn     int32  // number of documents to return
	CursorID           int64  // cursorID from the OpReply
}
