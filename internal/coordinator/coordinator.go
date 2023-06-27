package coordinator

import (
	"bufio"
	"log"
	"net"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/namelew/DHTWithMutualExclusion/packages/messages"
	"github.com/namelew/DHTWithMutualExclusion/packages/queue"
)

const protocol = "tcp"

type Coordinator struct {
	request        *queue.Queue[messages.Message]
	mutex          *sync.Mutex
	freeRegion     chan interface{}
	criticalRegion bool
}

func Build() *Coordinator {
	return &Coordinator{
		request:        &queue.Queue[messages.Message]{},
		mutex:          &sync.Mutex{},
		freeRegion:     make(chan interface{}, 1),
		criticalRegion: false,
	}
}

func (cd *Coordinator) queueHandler() {
	for {
		<-cd.freeRegion
		cd.mutex.Lock()

		if cd.criticalRegion || cd.request.Empty() {
			cd.mutex.Unlock()
			continue
		}

		m := cd.request.Dequeue()

		go func(m *messages.Message) {
			conn, err := net.Dial(protocol, m.Lockback)

			if err != nil {
				log.Println("Unable to create connection with client ", m.Id, ". ", err.Error())
				return
			}

			log.Println("Allowing acess to ", m.Id)

			Send(conn, &messages.Message{
				Id:       0,
				Action:   messages.ALLOW,
				Lockback: os.Getenv("CTRADRESS"),
			})
		}(&m)

		cd.mutex.Unlock()
	}
}

func (cd *Coordinator) Handler() {
	godotenv.Load(".env")

	l, err := net.Listen(protocol, os.Getenv("CTRADRESS"))

	if err != nil {
		log.Panic("Unable to create lisntener. ", err.Error())
	}

	go cd.queueHandler()

	for {
		request, err := l.Accept()

		if err != nil {
			log.Println("Unable to serve connection. ", err.Error())
			continue
		}

		go func(c net.Conn) {
			var in, out messages.Message
			b := make([]byte, 1024)

			n, err := bufio.NewReader(c).Read(b)

			defer c.Close()

			if err != nil {
				log.Println("Unable to read data from "+c.RemoteAddr().String()+". ", err.Error())
				return
			}

			if err := in.Unpack(b[:n]); err != nil {
				log.Println("Unable to unpack data from "+c.RemoteAddr().String()+". ", err.Error())
				return
			}

			switch in.Action {
			case messages.REQUEST:
				cd.mutex.Lock()
				defer cd.mutex.Unlock()

				if !cd.criticalRegion {
					cd.criticalRegion = true
					log.Println(in.Id, "allowed to access critical region")
					out.Action = messages.ALLOW
				} else {
					log.Println(in.Id, "not allowed to access critical region")
					cd.request.Enqueue(in)
					out.Action = messages.REFUSE
				}
			case messages.FREE:
				cd.mutex.Lock()
				defer cd.mutex.Unlock()
				cd.criticalRegion = false
				out.Action = messages.ACKFREE
				log.Println(in.Id, "finished to use Critical region")
				cd.freeRegion <- true
			}

			Send(c, &out)
		}(request)
	}
}

func Send(c net.Conn, m *messages.Message) {
	payload, err := m.Pack()

	if err != nil {
		log.Println("Unable to pack message. ", err.Error())
		return
	}

	_, err = c.Write(payload)

	if err != nil {
		log.Println("Unable to send message to "+c.RemoteAddr().String()+". ", err.Error())
		return
	}
}
