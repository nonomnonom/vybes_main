package domain

// Counter helps in creating an auto-incrementing sequence.
type Counter struct {
	ID            string `bson:"_id"`
	SequenceValue int64  `bson:"sequence_value"`
}
