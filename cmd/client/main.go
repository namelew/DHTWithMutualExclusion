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

		var request, response messages.Message

		a, err := strconv.Atoi(input[0])

		if err != nil {
			log.Println(err.Error())
			continue
		}

		request.Action = messages.Action(a)
		request.Payload.CPF = input[2]
		request.Payload.Curso = sanitaze(input[3])
		request.Payload.Nome = sanitaze(strings.Join(input[4:], " "))

		fmt.Println("Expect: Turma1 Turma2 ...")

		p, err = r.ReadSlice('\n')

		if err != nil {
			log.Println(err.Error())
			continue
		}

		for _, turma := range strings.Split(string(p), " ") {
			request.Payload.Turmas = append(request.Payload.Turmas, sanitaze(turma))
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

		if err := request.Send(conn); err != nil {
			conn.Close()
			mutex.Unlock()
			log.Println("Unable to send request. ", err.Error())
			continue
		}

		if err := response.Receive(conn); err != nil {
			conn.Close()
			mutex.Unlock()
			log.Println("Unable to receive response. ", err.Error())
			continue
		}

		conn.Close()

		mutex.Unlock()

		switch response.Action {
		case messages.ACK:
			log.Printf("Sucess! ")
			switch request.Action {
			case messages.INSERT:
				log.Printf("Registro de chave %s foi inserido corretamente na DHT\n", request.Payload.Value())
			case messages.QUERY:
				log.Println("Resultado:")
				log.Println("Nome:", response.Payload.Nome)
				log.Println("CPF:", response.Payload.CPF)
				log.Println("Curso:", response.Payload.Curso)
				log.Println("Turmas:", response.Payload.Turmas)
			case messages.REMOVE:
				log.Printf("Registro de chave %s foi removido corretamente da DHT\n", request.Payload.Value())
			default:
				log.Println("Resposta inesperada!")
			}
		case messages.ERROR:
			log.Println("Erro!")
			switch request.Action {
			case messages.INSERT:
				log.Printf("Error ao inserir registro de chave %s na DHT\n", request.Payload.Value())
			case messages.QUERY:
				log.Printf("Error ao busca registro de chave %s na DHT\n", request.Payload.Value())
			case messages.REMOVE:
				log.Printf("Error ao remover registro de chave %s da DHT\n", request.Payload.Value())
			default:
				log.Println("Error inesperado!")
			}
		}
	}
}
