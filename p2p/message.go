package p2p

import (
	"io"
	"myeth/rlp"
	"time"
)

//协议消息 读写的接口
type MsgReader interface {
	ReadMsg() (Msg, error)
}

type MsgWriter interface {
	// WriteMsg sends a message. It will block until the message's
	// Payload has been consumed by the other end.
	//
	// Note that messages can be sent only once because their
	// payload reader is drained.
	WriteMsg(Msg) error
}

//协议读写接口
type MsgReadWriter interface {
	MsgReader
	MsgWriter
}

// Msg defines the structure of a p2p message.
//
// Note that a Msg can only be sent once since the Payload reader is
// consumed during sending. It is not possible to create a Msg and
// send it any number of times. If you want to reuse an encoded
// structure, encode the payload into a byte array and create a
// separate Msg with a bytes.Reader as Payload for each send.
// 一个p2p消息的定义
type Msg struct {
	Code       uint64
	Size       uint32 // size of the paylod
	Payload    io.Reader
	ReceivedAt time.Time
}

//NEED DO!! 这个函数的功能还有待考究
func (msg Msg) Discard() error {
	return nil
}

//发送一个消息结构体 使用 w接口  数据使用RLP encoded的
//w 是 protoRW, msgcode 是发送的协议 命令字, data是发送的数据
func Send(w MsgWriter, msgcode uint64, data interface{}) error {
	size, r, err := rlp.EncodeToReader(data)
	if err != nil {
		return err
	}
	return w.WriteMsg(Msg{Code: msgcode, Size: uint32(size), Payload: r})
}

//支持发送 0 到 多个数据
func SendItems(w MsgWriter, msgcode uint64, elems ...interface{}) error {
	return Send(w, msgcode, elems)
}

func (msg Msg) Decode(val interface{}) error {
	s := rlp.NewStream(msg.Payload, uint64(msg.Size))
	if err := s.Decode(val); err != nil {
		return newPeerError(errInvalidMsg, "(code %x) (size %d) %v", msg.Code, msg.Size, err)
	}
	return nil
}
