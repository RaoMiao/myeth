package keystore

import (
	crand "crypto/rand"
	"myeth/accounts"
)

type KeyStore struct {
}

func NewKeyStore() *KeyStore {
	ks := &KeyStore{}
	return ks
}

func (ks *KeyStore) NewAccount(passphrase string) (accounts.Account, error) {
	_, account, err := storeNewKey(crand.Reader, passphrase)
	if err != nil {
		return accounts.Account{}, err
	}
	return account, nil
}
