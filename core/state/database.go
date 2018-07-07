package state

import (
	"myeth/common"
	"myeth/ethdb"
)

type Database interface {
	// OpenTrie opens the main account trie.
	OpenTrie(root common.Hash) (Trie, error)

	// OpenStorageTrie opens the storage trie of an account.
	OpenStorageTrie(addrHash, root common.Hash) (Trie, error)
}

type cachingDB struct {
}

//Merkle Trie
type Trie interface {
	TryGet(key []byte) ([]byte, error)
	TryUpdate(key, value []byte) error
	TryDelete(key []byte) error
	//Commit(onleaf trie.LeafCallback) (common.Hash, error)
	Hash() common.Hash
}

func NewDatabase(db ethdb.Database) Database {
	return &cachingDB{}
}
