package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/namelew/DHTWithMutualExclusion/packages/messages"
)

func sanitaze(s string) string {
	trash := []string{"\n", "\b", "\r", "\t"}

	for i := range trash {
		s = strings.ReplaceAll(s, trash[i], "")
	}

	return s
}

func main() {
	r := bufio.NewReader(os.Stdin)

	for {
		p, err := r.ReadSlice('\n')

		if err != nil {
			log.Println(err.Error())
			continue
		}

		input := strings.Split(string(p), " ")

		if len(input) < 4 {
			continue
		}

		var m messages.Message

		a, err := strconv.Atoi(input[0])

		if err != nil {
			log.Println(err.Error())
			continue
		}

		m.Action = messages.Action(a)
		m.Key = input[2]
		m.Name = sanitaze(strings.Join(input[3:], " "))

		conn, err := net.Dial("tcp", input[1])

		if err != nil {
			log.Println(err.Error())
			continue
		}

		if err := m.Send(conn); err != nil {
			conn.Close()
			log.Println("Unable to send request. ", err.Error())
			continue
		}

		if err := m.Receive(conn); err != nil {
			conn.Close()
			log.Println("Unable to receive response. ", err.Error())
			continue
		}

		conn.Close()

		log.Println(m)
	}
}
