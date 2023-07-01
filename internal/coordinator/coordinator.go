package coordinator

import (
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/namelew/DHTWithMutualExclusion/packages/messages"
	"github.com/namelew/DHTWithMutualExclusion/packages/queue"
)

const protocol = "tcp"

type Coordinator struct {
	request        *queue.Queue[messages.Message]
	mutex          *sync.Mutex
	freeRegion     chan bool
	criticalRegion bool
}

func Build() *Coordinator {
	return &Coordinator{
		request:        &queue.Queue[messages.Message]{},
		mutex:          &sync.Mutex{},
		freeRegion:     make(chan bool, 1),
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
				log.Println("Unable to create connection with client", m.Id, ".", err.Error())
				return
			}

			defer conn.Close()

			log.Println("Allowing acess to", m.Id)

			response := messages.Message{
				Id:       0,
				Action:   messages.ALLOW,
				Lockback: os.Getenv("CTRADRESS"),
			}

			if err := response.Send(conn); err != nil {
				log.Println("Queue Handler: can't send allow message.", err.Error())
				return
			}
		}(&m)

		cd.mutex.Unlock()
	}
}

func (cd *Coordinator) Handler() {
	if err := godotenv.Load(".env"); err != nil {
		log.Panic(err.Error())
	}

	l, err := net.Listen(protocol, os.Getenv("CTRADRESS"))

	if err != nil {
		log.Panic("Unable to create lisntener. ", err.Error())
	}

	go cd.queueHandler()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Finishing Coordinator")
		os.Exit(0)
	}()

	for {
		request, err := l.Accept()

		if err != nil {
			log.Println("Unable to serve connection. ", err.Error())
			continue
		}

		go func(c net.Conn) {
			var in, out messages.Message

			defer c.Close()

			if err := in.Receive(c); err != nil {
				log.Println("Handler Request error:", err.Error())
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

			if err := out.Send(c); err != nil {
				log.Println("Handler request error:", err.Error())
				return
			}
		}(request)
	}
}
