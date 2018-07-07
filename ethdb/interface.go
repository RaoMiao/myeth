package ethdb

type Putter interface {
	Put(key []byte, value []byte) error
}

type Database interface {
	Putter
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Delete(key []byte) error
	Close()
	//NewBatch() Batch
}

type Batch interface {
	Putter
	ValueSize() int
	Write() error
	Reset()
}