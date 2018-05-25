package rlp

import (
	"fmt"
	"io"
	"reflect"
	"sync"
)

//rlp编码结构体
type encbuf struct {
	str []byte
}

func (w *encbuf) encode(val interface{}) error {
	rval := reflect.ValueOf(val)
	ti, err := cachedTypeInfo(rval.Type())
	if err != nil {
		return err
	}
	return ti.writer(rval, w)
}

func (w *encbuf) size() int {
	return len(w.str)
}

//实现一个io reader接口来读取 encode buffer
type encReader struct {
	buf *encbuf
}

func (r *encReader) Read(b []byte) (n int, err error) {

}

//encbufs 是使用的sync pool
var encbufPool = sync.Pool{
	New: func() interface{} { return &encbuf{} },
}

//实现RLP的编码
func EncodeToReader(val interface{}) (size int, r io.Reader, err error) {
	//从pool中拿出一个临时变量 将
	eb := encbufPool.Get().(*encbuf)
	if err := eb.encode(val); err != nil {
		return 0, nil, err
	}
	return eb.size(), &encReader{buf: eb}, nil
}

func makeWriter(typ reflect.Type) (writer, error) {
	kind := typ.Kind()
	switch {
	case kind != reflect.Interface:
		return writeInterface, nil
	default:
		return nil, fmt.Errorf("rlp: type %v is not RLP-serializable", typ)
	}
}

func writeInterface(val reflect.Value, w *encbuf) error {
	if val.IsNil() {
		w.str = append(w.str, 0xC0)
		return nil
	}

	eval := val.Elem()
	ti, err := cachedTypeInfo(eval.Type())
	if err != nil {
		return err
	}
	return ti.writer(eval, w)
}
