// Package msg is 100% generated. If you edit this file,
// you will lose your changes at the next build cycle.
// DO NOT MAKE ANY CHANGES YOU WISH TO KEEP.
//
// The correct places for commits are:
//  - The XML model used for this code generation: zre_msg.xml
//  - The code generation script that built this file: zproto_codec_go
package msg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	zmq "github.com/pebbe/zmq4"
)

const (
	// Signature is put into every protocol message and lets us filter bogus
	// or unknown protocols. It is a 4-bit number from 0 to 15. Use a unique value
	// for each protocol you write, at least.
	Signature uint16 = 0xAAA0 | 0
)

// Definition of message IDs
const (
	HelloID   uint8 = 1
	WhisperID uint8 = 2
	ShoutID   uint8 = 3
	JoinID    uint8 = 4
	LeaveID   uint8 = 5
	PingID    uint8 = 6
	PingOkID  uint8 = 7
)

// Transit is a codec interface
type Transit interface {
	Marshal() ([]byte, error)
	Unmarshal(...[]byte) error
	String() string
	Send(*zmq.Socket) error
	SetRoutingID([]byte)
	RoutingID() []byte
	SetVersion(byte)
	Version() byte
	SetSequence(uint16)
	Sequence() uint16
}

// Unmarshal unmarshals data from raw frames.
func Unmarshal(frames ...[]byte) (t Transit, err error) {
	if frames == nil {
		return nil, errors.New("can't unmarshal an empty message")
	}
	var buffer *bytes.Buffer

	// Check the signature
	var signature uint16
	buffer = bytes.NewBuffer(frames[0])
	binary.Read(buffer, binary.BigEndian, &signature)
	if signature != Signature {
		// Invalid signature
		return nil, fmt.Errorf("invalid signature %X != %X", Signature, signature)
	}

	// Get message id and parse per message type
	var id uint8
	binary.Read(buffer, binary.BigEndian, &id)

	switch id {
	case HelloID:
		t = NewHello()
	case WhisperID:
		t = NewWhisper()
	case ShoutID:
		t = NewShout()
	case JoinID:
		t = NewJoin()
	case LeaveID:
		t = NewLeave()
	case PingID:
		t = NewPing()
	case PingOkID:
		t = NewPingOk()
	}
	err = t.Unmarshal(frames...)

	return t, err
}

// Recv receives marshaled data from a 0mq socket.
func Recv(socket *zmq.Socket) (t Transit, err error) {
	return recv(socket, 0)
}

// RecvNoWait receives marshaled data from 0mq socket. It won't wait for input.
func RecvNoWait(socket *zmq.Socket) (t Transit, err error) {
	return recv(socket, zmq.DONTWAIT)
}

// recv receives marshaled data from 0mq socket.
func recv(socket *zmq.Socket, flag zmq.Flag) (t Transit, err error) {
	// Read all frames
	frames, err := socket.RecvMessageBytes(flag)
	if err != nil {
		return nil, err
	}

	sType, err := socket.GetType()
	if err != nil {
		return nil, err
	}

	var routingID []byte
	// If message came from a router socket, first frame is routingID
	if sType == zmq.ROUTER {
		if len(frames) <= 1 {
			return nil, errors.New("no routingID")
		}
		routingID = frames[0]
		frames = frames[1:]
	}

	t, err = Unmarshal(frames...)
	if err != nil {
		return nil, err
	}

	if sType == zmq.ROUTER {
		t.SetRoutingID(routingID)
	}
	return t, err
}

// Clone clones a message.
func Clone(t Transit) Transit {

	switch msg := t.(type) {
	case *Hello:
		cloned := NewHello()
		routingID := make([]byte, len(msg.RoutingID()))
		copy(routingID, msg.RoutingID())
		cloned.SetRoutingID(routingID)
		cloned.version = msg.version
		cloned.sequence = msg.sequence
		cloned.Endpoint = msg.Endpoint
		for idx, str := range msg.Groups {
			cloned.Groups[idx] = str
		}
		cloned.Status = msg.Status
		cloned.Name = msg.Name
		for key, val := range msg.Headers {
			cloned.Headers[key] = val
		}
		return cloned

	case *Whisper:
		cloned := NewWhisper()
		routingID := make([]byte, len(msg.RoutingID()))
		copy(routingID, msg.RoutingID())
		cloned.SetRoutingID(routingID)
		cloned.version = msg.version
		cloned.sequence = msg.sequence
		cloned.Content = append(cloned.Content, msg.Content...)
		return cloned

	case *Shout:
		cloned := NewShout()
		routingID := make([]byte, len(msg.RoutingID()))
		copy(routingID, msg.RoutingID())
		cloned.SetRoutingID(routingID)
		cloned.version = msg.version
		cloned.sequence = msg.sequence
		cloned.Group = msg.Group
		cloned.Content = append(cloned.Content, msg.Content...)
		return cloned

	case *Join:
		cloned := NewJoin()
		routingID := make([]byte, len(msg.RoutingID()))
		copy(routingID, msg.RoutingID())
		cloned.SetRoutingID(routingID)
		cloned.version = msg.version
		cloned.sequence = msg.sequence
		cloned.Group = msg.Group
		cloned.Status = msg.Status
		return cloned

	case *Leave:
		cloned := NewLeave()
		routingID := make([]byte, len(msg.RoutingID()))
		copy(routingID, msg.RoutingID())
		cloned.SetRoutingID(routingID)
		cloned.version = msg.version
		cloned.sequence = msg.sequence
		cloned.Group = msg.Group
		cloned.Status = msg.Status
		return cloned

	case *Ping:
		cloned := NewPing()
		routingID := make([]byte, len(msg.RoutingID()))
		copy(routingID, msg.RoutingID())
		cloned.SetRoutingID(routingID)
		cloned.version = msg.version
		cloned.sequence = msg.sequence
		return cloned

	case *PingOk:
		cloned := NewPingOk()
		routingID := make([]byte, len(msg.RoutingID()))
		copy(routingID, msg.RoutingID())
		cloned.SetRoutingID(routingID)
		cloned.version = msg.version
		cloned.sequence = msg.sequence
		return cloned
	}

	return nil
}

// putString marshals a string into the buffer.
func putString(buffer *bytes.Buffer, str string) {
	size := len(str)
	binary.Write(buffer, binary.BigEndian, byte(size))
	binary.Write(buffer, binary.BigEndian, []byte(str[0:size]))
}

// getString unmarshals a string from the buffer.
func getString(buffer *bytes.Buffer) string {
	var size byte
	binary.Read(buffer, binary.BigEndian, &size)
	str := make([]byte, size)
	binary.Read(buffer, binary.BigEndian, &str)
	return string(str)
}

// putLongString marshals a string into the buffer.
func putLongString(buffer *bytes.Buffer, str string) {
	size := len(str)
	binary.Write(buffer, binary.BigEndian, uint32(size))
	binary.Write(buffer, binary.BigEndian, []byte(str[0:size]))
}

// getLongString unmarshals a string from the buffer.
func getLongString(buffer *bytes.Buffer) string {
	var size uint32
	binary.Read(buffer, binary.BigEndian, &size)
	str := make([]byte, size)
	binary.Read(buffer, binary.BigEndian, &str)
	return string(str)
}

// putBytes marshals []byte into the buffer.
func putBytes(buffer *bytes.Buffer, data []byte) {
	size := uint64(len(data))
	binary.Write(buffer, binary.BigEndian, size)
	binary.Write(buffer, binary.BigEndian, data)
}

// getBytes unmarshals []byte from the buffer.
func getBytes(buffer *bytes.Buffer) []byte {
	var size uint64
	binary.Read(buffer, binary.BigEndian, &size)
	data := make([]byte, size)
	binary.Read(buffer, binary.BigEndian, &data)
	return data
}
