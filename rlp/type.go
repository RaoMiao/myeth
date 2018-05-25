package rlp

import (
	"reflect"
)

//
//type decoder func(*Stream, reflect.Value) error
type writer func(reflect.Value, *encbuf) error

type typeinfo struct {
	//decoder
	writer
}

func cachedTypeInfo(typ reflect.Type) (*typeinfo, error) {
	return cachedTypeInfo1(typ)
}

func cachedTypeInfo1(typ reflect.Type) (*typeinfo, error) {
	info, err := genTypeInfo(typ)
	return info, err
}

func genTypeInfo(typ reflect.Type) (info *typeinfo, err error) {
	info = new(typeinfo)

	//
	if info.writer, err = makeWriter(typ); err != nil {
		return nil, err
	}
	return info, nil
}
