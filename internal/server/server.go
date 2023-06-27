package server

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/namelew/DHTWithMutualExclusion/packages/hashtable"
	"github.com/namelew/DHTWithMutualExclusion/packages/messages"
)

type Node struct {
	adress string
	start  int
	end    int
}

type FileSystem struct {
	adress       string
	id           uint64
	start        int
	end          int
	lock         sync.Mutex
	neighborhood []Node
	inodes       hashtable.HashTable[string, string]
}

const SOURCEFILE = "./routing_table.in"

func removeBackSlash(s string) string {
	backslash := []string{"\n", "\a", "\b", "\r"}

	for i := range backslash {
		s = strings.ReplaceAll(s, backslash[i], "")
	}

	return s
}

func New(id uint64) *FileSystem {
	data, err := os.ReadFile(SOURCEFILE)

	if err != nil {
		log.Panic("Unable to create file system. Error on sourcefile read: ", err.Error())
	}

	lines := strings.Split(string(data), "\n")

	size, err := strconv.Atoi(removeBackSlash(lines[0]))

	if err != nil {
		log.Panic("Unable to create file system. Error on table size load: ", err.Error())
	}

	var start, end int = 0, 0
	table := make([]Node, 0)

	for i := range lines {
		if i > 0 {
			cols := strings.Split(lines[i], " ")

			if len(cols) < 4 {
				continue
			}

			server := Node{}

			server.adress = removeBackSlash(cols[1])

			start, err = strconv.Atoi(removeBackSlash(cols[2]))

			if err != nil {
				log.Panic("Unable to create file system. Error on table start load: ", err.Error())
			}

			end, err = strconv.Atoi(removeBackSlash(cols[3]))

			if err != nil {
				log.Panic("Unable to create file system. Error on table end load: ", err.Error())
			}

			server.start = start
			server.end = end

			table = append(table, server)
		}
	}

	return &FileSystem{
		id:           id,
		adress:       table[id].adress,
		start:        table[id].start,
		end:          table[id].end,
		neighborhood: table,
		inodes:       hashtable.New[string, string](&hashtable.Open[string, string]{}, hashtable.Common{Size: size, End: table[id].end}),
	}
}

func (fs *FileSystem) redirect(m *messages.Message) messages.Message {
	for nid := range fs.neighborhood {
		slot := fs.inodes.Hash(m)

		if slot >= fs.neighborhood[nid].start && slot <= fs.neighborhood[nid].end {
			log.Println("Request redirect to server", nid, " in adress", fs.neighborhood[nid].adress)
			conn, err := net.Dial("tcp", fs.neighborhood[nid].adress)

			if err != nil {
				log.Println("Unable to create connection with node", nid, ":", err.Error())
				return messages.Message{}
			}

			defer conn.Close()

			if err := m.Send(conn); err != nil {
				log.Println("Unable to send request. ", err.Error())
				return messages.Message{}
			}

			if err := m.Receive(conn); err != nil {
				log.Println("Unable to receive response. ", err.Error())
				continue
			}

			return *m
		}
	}

	return messages.Message{}
}

func (fs *FileSystem) insert(m *messages.Message) messages.Message {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	if fs.inodes.Hash(m) >= fs.start && fs.inodes.Hash(m) <= fs.end {
		if err := fs.inodes.Insert(m, m.Name); err != nil {
			log.Printf("Unable to insert %s in register: %s\n", m.Name, err.Error())
			return messages.Message{}
		}
		log.Printf("Register %s was inserted with key %s in slot %d\n", m.Name, m.Key, fs.inodes.Hash(m))
	} else {
		log.Println("Out of domain! Redirecting request...")
		return fs.redirect(m)
	}

	return messages.Message{Action: messages.ACK}
}

func (fs *FileSystem) query(m *messages.Message) messages.Message {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	response := messages.Message{
		Key:    m.Key,
		Action: messages.ACK,
	}

	if fs.inodes.Hash(m) >= fs.start && fs.inodes.Hash(m) <= fs.end {
		data, err := fs.inodes.Search(m)
		if err != nil {
			log.Printf("Unable to find %s in register: %s\n", m.Key, err.Error())
			return messages.Message{}
		}
		response.Name = data
	} else {
		log.Println("Out of domain! Redirecting request...")
		return fs.redirect(m)
	}

	return response
}

func (fs *FileSystem) remove(m *messages.Message) messages.Message {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	if fs.inodes.Hash(m) >= fs.start && fs.inodes.Hash(m) <= fs.end {
		if err := fs.inodes.Delete(m); err != nil {
			log.Printf("Unable to remove %s from register: %s\n", m.Name, err.Error())
			return messages.Message{}
		}
		log.Printf("Register in key %s was removed\n", m.Key)
	} else {
		log.Println("Out of domain! Redirecting request...")
		return fs.redirect(m)
	}

	return messages.Message{Action: messages.ACK}
}

func (fs *FileSystem) handlerRequests() {
	listener, err := net.Listen("tcp", fs.adress)

	if err != nil {
		log.Panic("Unable to create request handler: ", err.Error())
	}

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Enable to create connection: ", err.Error())
			continue
		}

		go func(c net.Conn) {
			var request, response messages.Message

			if err := request.Receive(c); err != nil {
				log.Println("Unable to receive message from ", c.RemoteAddr().String(), ":", err.Error())
				return
			}

			switch request.Action {
			case messages.INSERT:
				response = fs.insert(&request)
			case messages.QUERY:
				response = fs.query(&request)
			case messages.REMOVE:
				response = fs.remove(&request)
			}

			if err := response.Send(c); err != nil {
				log.Println("Unable to send response to ", c.RemoteAddr().String(), ":", err.Error())
			}
		}(conn)
	}
}

func (fs *FileSystem) Build() {
	log.Println("Init request handler...")
	go fs.handlerRequests()

	log.Printf("File System Node %d started\n", fs.id)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
