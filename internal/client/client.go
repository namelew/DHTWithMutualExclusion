package client

import (
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/namelew/DHTWithMutualExclusion/packages/messages"
)

const protocol = "tcp"

type Client struct {
	id      int
	adress  string
}

func New(id int, adress string) *Client {
	return &Client{
		id:      id,
		adress:  adress,
	}
}

func (c *Client) Run() {
	for {
		switch c.Lock() {
		case messages.ALLOW:
		case messages.REFUSE:
			c.Wait()
		default:
			log.Println("Permition error: unable to acess DHT")
		}
	}
}

func (c *Client) Wait() {
	waitTime := rand.Intn(10)
	l, err := net.Listen("tcp", c.adress)

	if err != nil {
		log.Println("Error to open listener.", err.Error())
		time.Sleep(time.Second * time.Duration(waitTime))
		return
	}

	for {
		c, err := l.Accept()

		if err != nil {
			log.Println("Error to accept connection.", err.Error())
			time.Sleep(time.Second * time.Duration(waitTime))
			continue
		}

		request := messages.Message{}

		if err := request.Receive(c); err != nil {
			log.Println("Error to accept connection.", err.Error())
			time.Sleep(time.Second * time.Duration(waitTime))
			continue
		}

		if request.Action == messages.ALLOW {
			l.Close()
			return
		}
	}
}

func (c *Client) Lock() messages.Action {
	conn, err := net.Dial(protocol, os.Getenv("CTRADRESS"))

	if err != nil {
		log.Println("Unable to create connection with Coordenator.", err.Error())
		return messages.ERROR
	}

	request := messages.Message{
		Id:       uint64(c.id),
		Action:   messages.REQUEST,
		Lockback: c.adress,
	}

	if err := request.Send(conn); err != nil {
		log.Println("Unable to send request from coordinator.", err.Error())
		return messages.ERROR
	}

	var response messages.Message

	if err := response.Receive(conn); err != nil {
		log.Println("Unable to read data from coordinator.", err.Error())
		return messages.ERROR
	}

	return response.Action
}

func (c *Client) Unlock() {
	conn, err := net.Dial(protocol, os.Getenv("CTRADRESS"))

	if err != nil {
		log.Println("Unable to create connection with Coordenator.", err.Error())
		return
	}

	request := messages.Message{
		Id:       uint64(c.id),
		Action:   messages.FREE,
		Lockback: c.adress,
	}

	if err := request.Receive(conn); err != nil {
		log.Println("Unable to send unlock message.", err.Error())
		return
	}
}
