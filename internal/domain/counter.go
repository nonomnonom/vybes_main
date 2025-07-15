package domain

// Counter helps in creating an auto-incrementing sequence.
type Counter struct {
	ID  string `bson:"_id"`
	Seq int64  `bson:"seq"`
}
