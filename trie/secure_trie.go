package trie

import "myeth/common"

type SecureTrie struct {
	trie Trie
}

func NewSecure(root common.Hash, db *Database, cachelimit uint16) (*SecureTrie, error) {

}
