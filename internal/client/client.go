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
	usetime time.Duration
}

func New(id int, adress string, useDelay time.Duration) *Client {
	return &Client{
		id:      id,
		adress:  adress,
		usetime: useDelay,
	}
}

func (c *Client) Run() {
	enterRegion := func() {
		log.Printf("Client %d enter the critical region\n", c.id)
		time.Sleep(c.usetime)
		c.Unlock()
	}
	for {
		switch c.Lock() {
		case messages.ALLOW:
			enterRegion()
		case messages.REFUSE:
			waitTime := rand.Intn(10)
			l, err := net.Listen("tcp", c.adress)

			if err != nil {
				log.Println("Error to open listener. ", err.Error())
				time.Sleep(time.Second * time.Duration(waitTime))
				continue
			}

			for {
				c, err := l.Accept()

				if err != nil {
					log.Println("Error to accept connection. ", err.Error())
					time.Sleep(time.Second * time.Duration(waitTime))
					continue
				}

				buffer := make([]byte, 1024)
				var m messages.Message
				n, err := c.Read(buffer)

				if err != nil {
					log.Println("Error to read from connection. ", err.Error())
					c.Close()
					time.Sleep(time.Second * time.Duration(waitTime))
					continue
				}

				c.Close()

				if err := m.Unpack(buffer[:n]); err != nil {
					log.Println("Unable to unpack data from buffer. ", err.Error())
					time.Sleep(time.Second * time.Duration(waitTime))
					continue
				}

				if m.Action == messages.ALLOW {
					enterRegion()
					break
				}
			}

			l.Close()
		default:
			waitTime := rand.Intn(10)
			log.Printf("Client %d can't enter the critical region and will sleep %d seconds\n", c.id, waitTime)
			time.Sleep(time.Second * time.Duration(waitTime))
		}
	}
}

func (c *Client) Lock() messages.Action {
	conn, err := net.Dial(protocol, os.Getenv("CTRADRESS"))
	buffer := make([]byte, 256)

	if err != nil {
		log.Println("Unable to create connection with Coordenator. ", err.Error())
		return messages.ERROR
	}

	request := messages.Message{
		Id:       uint64(c.id),
		Action:   messages.REQUEST,
		Lockback: c.adress,
	}

	requestPayload, err := request.Pack()

	if err != nil {
		log.Println("Unable to create request message. ", err.Error())
		return messages.ERROR
	}

	conn.Write(requestPayload)

	time.After(time.Second * 5)

	n, err := conn.Read(buffer)

	if err != nil {
		log.Println("Unable to read data from coordinator. ", err.Error())
		return messages.ERROR
	}

	response := messages.Message{}

	response.Unpack(buffer[:n])

	return response.Action
}

func (c *Client) Unlock() {
	conn, err := net.Dial(protocol, os.Getenv("CTRADRESS"))

	if err != nil {
		log.Println("Unable to create connection with Coordenator. ", err.Error())
		return
	}

	request := messages.Message{
		Id:       uint64(c.id),
		Action:   messages.FREE,
		Lockback: c.adress,
	}

	requestPayload, err := request.Pack()

	if err != nil {
		log.Println("Unable to create unlock message. ", err.Error())
		return
	}

	conn.Write(requestPayload)

	time.After(time.Second * 5)
}
