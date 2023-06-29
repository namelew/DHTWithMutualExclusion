package messages

import (
	"encoding/json"
	"errors"
	"net"
)

type Action uint32

const (
	ERROR   Action = 0
	INSERT  Action = 1
	REMOVE  Action = 2
	QUERY   Action = 3
	REQUEST Action = 4
	ALLOW   Action = 5
	REFUSE  Action = 6
	FREE    Action = 7
	ACKFREE Action = 8
	ACK     Action = 9
)

type Aluno struct {
	CPF    string
	Nome   string
	Curso  string
	Turmas []string
}

type Message struct {
	Id       uint64
	Action   Action
	Payload  Aluno
	Lockback string
}

func (a *Aluno) Index() int {
	sum := 0

	for i := range a.CPF {
		sum += int(a.CPF[i])
	}

	return sum
}

func (a *Aluno) Value() string {
	return a.CPF
}

func (m *Message) Pack() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Unpack(b []byte) error {
	return json.Unmarshal(b, m)
}

func (m *Message) Send(c net.Conn) error {
	payload, err := m.Pack()

	if err != nil {
		return errors.New("Unable to pack message. " + err.Error())
	}

	_, err = c.Write(payload)

	if err != nil {
		return errors.New("Unable to send message to " + c.RemoteAddr().String() + ". " + err.Error())
	}

	return err
}

func (m *Message) Receive(c net.Conn) error {
	buffer := make([]byte, 1024)

	n, err := c.Read(buffer)

	if err != nil {
		return errors.New("Unable to receive message from " + c.RemoteAddr().String() + ". " + err.Error())
	}

	if err := m.Unpack(buffer[:n]); err != nil {
		return errors.New("Unable to read message from " + c.RemoteAddr().String() + ". " + err.Error())
	}

	return err
}
