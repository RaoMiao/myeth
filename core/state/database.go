package state

import (
	"myeth/common"
	"myeth/ethdb"
	"myeth/trie"
	"sync"
)

type Database interface {
	// OpenTrie opens the main account trie.
	OpenTrie(root common.Hash) (Trie, error)

	// OpenStorageTrie opens the storage trie of an account.
	OpenStorageTrie(addrHash, root common.Hash) (Trie, error)
}

type cachingDB struct {
	db *trie.Database
	mu sync.Mutex
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
	return &cachingDB{
		db: trie.NewDatabase(db),
	}
}

func (db *cachingDB) OpenTrie(root common.Hash) (Trie, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	tr, err := trie.Ne
}

// cachedTrie inserts its trie into a cachingDB on commit.
type cachedTrie struct {
	*trie.SecureTrie
	db *cachingDB
}
