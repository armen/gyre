package msg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	zmq "github.com/pebbe/zmq4"
)

// Leave struct
// Leave a group
type Leave struct {
	routingID []byte
	version   byte
	sequence  uint16
	Group     string
	Status    byte
}

// NewLeave creates new Leave message.
func NewLeave() *Leave {
	leave := &Leave{}
	return leave
}

// String returns print friendly name.
func (l *Leave) String() string {
	str := "ZRE_MSG_LEAVE:\n"
	str += fmt.Sprintf("    version = %v\n", l.version)
	str += fmt.Sprintf("    sequence = %v\n", l.sequence)
	str += fmt.Sprintf("    Group = %v\n", l.Group)
	str += fmt.Sprintf("    Status = %v\n", l.Status)
	return str
}

// Marshal serializes the message.
func (l *Leave) Marshal() ([]byte, error) {
	// Calculate size of serialized data
	bufferSize := 2 + 1 // Signature and message ID

	// version is a 1-byte integer
	bufferSize++

	// sequence is a 2-byte integer
	bufferSize += 2

	// Group is a string with 1-byte length
	bufferSize++ // Size is one byte
	bufferSize += len(l.Group)

	// Status is a 1-byte integer
	bufferSize++

	// Now serialize the message
	tmpBuf := make([]byte, bufferSize)
	tmpBuf = tmpBuf[:0]
	buffer := bytes.NewBuffer(tmpBuf)
	binary.Write(buffer, binary.BigEndian, Signature)
	binary.Write(buffer, binary.BigEndian, LeaveID)

	// version
	value, _ := strconv.ParseUint("2", 10, 1*8)
	binary.Write(buffer, binary.BigEndian, byte(value))

	// sequence
	binary.Write(buffer, binary.BigEndian, l.sequence)

	// Group
	putString(buffer, l.Group)

	// Status
	binary.Write(buffer, binary.BigEndian, l.Status)

	return buffer.Bytes(), nil
}

// Unmarshal unmarshals the message.
func (l *Leave) Unmarshal(frames ...[]byte) error {
	if frames == nil {
		return errors.New("Can't unmarshal empty message")
	}

	frame := frames[0]
	frames = frames[1:]

	buffer := bytes.NewBuffer(frame)

	// Get and check protocol signature
	var signature uint16
	binary.Read(buffer, binary.BigEndian, &signature)
	if signature != Signature {
		return fmt.Errorf("invalid signature %X != %X", Signature, signature)
	}

	// Get message id and parse per message type
	var id uint8
	binary.Read(buffer, binary.BigEndian, &id)
	if id != LeaveID {
		return errors.New("malformed Leave message")
	}
	// version
	binary.Read(buffer, binary.BigEndian, &l.version)
	if l.version != 2 {
		return errors.New("malformed version message")
	}
	// sequence
	binary.Read(buffer, binary.BigEndian, &l.sequence)
	// Group
	l.Group = getString(buffer)
	// Status
	binary.Read(buffer, binary.BigEndian, &l.Status)

	return nil
}

// Send sends marshaled data through 0mq socket.
func (l *Leave) Send(socket *zmq.Socket) (err error) {
	frame, err := l.Marshal()
	if err != nil {
		return err
	}

	socType, err := socket.GetType()
	if err != nil {
		return err
	}

	// If we're sending to a ROUTER, we send the routingID first
	if socType == zmq.ROUTER {
		_, err = socket.SendBytes(l.routingID, zmq.SNDMORE)
		if err != nil {
			return err
		}
	}

	// Now send the data frame
	_, err = socket.SendBytes(frame, 0)
	if err != nil {
		return err
	}

	return err
}

// RoutingID returns the routingID for this message, routingID should be set
// whenever talking to a ROUTER.
func (l *Leave) RoutingID() []byte {
	return l.routingID
}

// SetRoutingID sets the routingID for this message, routingID should be set
// whenever talking to a ROUTER.
func (l *Leave) SetRoutingID(routingID []byte) {
	l.routingID = routingID
}

// SetVersion sets the version.
func (l *Leave) SetVersion(version byte) {
	l.version = version
}

// Version returns the version.
func (l *Leave) Version() byte {
	return l.version
}

// SetSequence sets the sequence.
func (l *Leave) SetSequence(sequence uint16) {
	l.sequence = sequence
}

// Sequence returns the sequence.
func (l *Leave) Sequence() uint16 {
	return l.sequence
}
