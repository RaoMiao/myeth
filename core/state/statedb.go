package state

import (
	"fmt"
	"math/big"
	"myeth/common"
	"myeth/rlp"
)

//StateDB 用来存储 和 Merkle trie相关的所有事情
type StateDB struct {
	db   Database
	trie Trie

	// This map holds 'live' objects, which will get modified while processing a state transition.
	// 这个map里面保存所有存在的object 会在执行交易的过程中发生改变
	stateObjects      map[common.Address]*stateObject
	stateObjectsDirty map[common.Address]struct{}

	journal *journal
}

func New(root common.Hash, db Database) (*StateDB, error) {
	return &StateDB{}, nil
}

//根据地址查找一个StateObject
func (self *StateDB) getStateObject(addr common.Address) (stateObject *stateObject) {
	if obj := self.stateObjects[addr]; obj != nil {
		if obj.deleted {
			return nil
		}

		return obj
	}

	//load from database
	enc, _ := self.trie.TryGet(addr[:])
	if len(enc) == 0 {
		//self.setError(err)
		return nil
	}

	var data Account
	if err := rlp.DecodeBytes(enc, &data); err != nil {
		return nil
	}

	obj := newObject(self, addr, data)
	self.setStateObject(obj)
	return obj
}

func (self *StateDB) setStateObject(object *stateObject) {
	self.stateObjects[object.address] = object
}

func (self *StateDB) GetOrNewStateObject(addr common.Address) *stateObject {
	stateObject := self.getStateObject(addr)
	if stateObject == nil || stateObject.deleted {
		stateObject, _ = self.createObject(addr)
	}
	return stateObject
}

func (self *StateDB) createObject(addr common.Address) (newobj, prev *stateObject) {
	prev = self.getStateObject(addr)
	newobj = newObject(self, addr, Account{})
	newobj.setNonce(0) //交易次数0
	if prev == nil {
		self.journal.append(createObjectChange{account: &addr})
	} else {
		self.journal.append(resetObjectChange{prev: prev})
	}
	self.setStateObject(newobj)
	return newobj, prev
}

// deleteStateObject removes the given object from the state trie.
func (self *StateDB) deleteStateObject(stateObject *stateObject) {
	stateObject.deleted = true
	addr := stateObject.address
	self.trie.TryDelete(addr[:])
}

func (self *StateDB) updateStateObject(stateObject *stateObject) {
	addr := stateObject.address
	data, err := rlp.EncodeToBytes(stateObject)
	if err != nil {
		panic(fmt.Errorf("can't encode object at %x: %v", addr[:], err))
	}
	self.trie.TryUpdate(addr[:], data)
}

func (self *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
	}
}

func (self *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
	}
}

// func (self *StateDB) SetCode(addr common.Address, code []byte) {
// 	stateObject := self.GetOrNewStateObject(addr)
// 	if stateObject != nil {
// 		stateObject.SetCode(crypto.Keccak256Hash(code), code)
// 	}
// }

// func (self *StateDB) SetState(addr common.Address, key, value common.Hash) {
// 	stateObject := self.GetOrNewStateObject(addr)
// 	if stateObject != nil {
// 		stateObject.SetState(self.db, key, value)
// 	}
// }

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func (s *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	s.Finalise(deleteEmptyObjects)
	return s.trie.Hash()
}

//定稿
func (s *StateDB) Finalise(deleteEmptyObjects bool) {
	for addr := range s.journal.dirties {
		stateObject, exist := s.stateObjects[addr]
		if !exist {
			continue
		}

		if stateObject.suicided || (deleteEmptyObjects && stateObject.empty()) {
			s.deleteStateObject(stateObject)
		} else {
			stateObject.updateRoot(s.db)
			s.updateStateObject(stateObject)
		}
		s.stateObjectsDirty[addr] = struct{}{}
	}
}
