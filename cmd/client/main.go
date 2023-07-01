package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/namelew/DHTWithMutualExclusion/internal/mutex"
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
	if err := godotenv.Load(".env"); err != nil {
		log.Panic(err.Error())
	}

	r := bufio.NewReader(os.Stdin)

	mutex := mutex.New(3, "localhost:30010")

	for {
		fmt.Println("Expect: Action Adress Key Course Name")
		p, err := r.ReadSlice('\n')

		if err != nil {
			log.Println(err.Error())
			continue
		}

		input := strings.Split(string(p), " ")

		if len(input) < 6 {
			continue
		}

		var m messages.Message

		a, err := strconv.Atoi(input[0])

		if err != nil {
			log.Println(err.Error())
			continue
		}

		m.Action = messages.Action(a)
		m.Payload.CPF = input[2]
		m.Payload.Curso = sanitaze(input[3])
		m.Payload.Nome = sanitaze(strings.Join(input[4:], " "))

		fmt.Println("Expect: Turma1 Turma2 ...")

		p, err = r.ReadSlice('\n')

		if err != nil {
			log.Println(err.Error())
			continue
		}

		for _, turma := range strings.Split(string(p), " ") {
			m.Payload.Turmas = append(m.Payload.Turmas, sanitaze(turma))
		}

		if err := mutex.RequestAcess(); err != nil {
			log.Println(err.Error())
			continue
		}

		conn, err := net.Dial("tcp", input[1])

		if err != nil {
			mutex.Unlock()
			log.Println(err.Error())
			continue
		}

		if err := m.Send(conn); err != nil {
			conn.Close()
			mutex.Unlock()
			log.Println("Unable to send request. ", err.Error())
			continue
		}

		if err := m.Receive(conn); err != nil {
			conn.Close()
			mutex.Unlock()
			log.Println("Unable to receive response. ", err.Error())
			continue
		}

		conn.Close()

		mutex.Unlock()

		switch m.Action {
		case messages.ACK:
			empty := messages.Aluno{}
			log.Println("Sucess!")
			if m.Payload.Nome != empty.Nome {
				log.Println("Resultado")
				log.Println("Nome:", m.Payload.Nome)
				log.Println("CPF:", m.Payload.CPF)
				log.Println("Curso:", m.Payload.Curso)
				log.Println("Turmas:", m.Payload.Turmas)
			}
		case messages.ERROR:
			log.Println("Erro!")
		}
	}
}
