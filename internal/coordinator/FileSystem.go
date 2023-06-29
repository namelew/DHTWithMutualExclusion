package coordinator

import (
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/namelew/DHTWithMutualExclusion/packages/messages"
)

const ROUTINGFILE = "./routing_table.in"

type Node struct {
	Id     uint64
	Weight uint64
	Adress string
}

type FileSystem struct {
	factor *rand.Rand
	nodes  []Node
}

func removeBackSlash(s string) string {
	backslash := []string{"\n", "\a", "\b", "\r"}

	for i := range backslash {
		s = strings.ReplaceAll(s, backslash[i], "")
	}

	return s
}

func New() *FileSystem {
	data, err := os.ReadFile(ROUTINGFILE)

	if err != nil {
		log.Panic("Unable to create file system. Error on sourcefile read: ", err.Error())
	}

	lines := strings.Split(string(data), "\n")

	nodes := make([]Node, 0)

	for i := range lines {
		if len(lines[i]) > 1 {
			cols := strings.Split(lines[i], " ")

			if len(cols) < 4 {
				continue
			}

			nodes = append(nodes, Node{
				Id:     uint64(i),
				Adress: removeBackSlash(cols[1]),
				Weight: 1,
			})
		}
	}

	return &FileSystem{
		nodes:  nodes,
		factor: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (fs *FileSystem) balance() int {
	return fs.factor.Int() % len(fs.nodes)
}

func (fs *FileSystem) Add(aluno messages.Aluno) error {
	request := messages.Message{
		Id:      0,
		Action:  messages.INSERT,
		Payload: aluno,
	}

	conn, err := net.Dial("tcp", fs.nodes[fs.balance()].Adress)

	if err != nil {
		log.Printf("Erro ao inserir %v no sistema de arquivos. Problema ao abrir conexão\n", aluno)
		return err
	}

	defer conn.Close()

	if err := request.Send(conn); err != nil {
		log.Printf("Erro ao inserir %v no sistema de arquivos. Problema ao enviar requisição\n", aluno)
		return err
	}

	if err := request.Receive(conn); err != nil {
		log.Printf("Erro ao inserir %v no sistema de arquivos. Problema ao receber respostas\n", aluno)
		return err
	}

	return nil
}

func (fs *FileSystem) Remove(key string) error {
	request := messages.Message{
		Id:     0,
		Action: messages.INSERT,
		Payload: messages.Aluno{
			CPF: key,
		},
	}

	conn, err := net.Dial("tcp", fs.nodes[fs.balance()].Adress)

	if err != nil {
		log.Printf("Erro ao remover %s do sistema de arquivos. Problema ao abrir conexão\n", key)
		return err
	}

	defer conn.Close()

	if err := request.Send(conn); err != nil {
		log.Printf("Erro ao remover %s do sistema de arquivos. Problema ao enviar requisição\n", key)
		return err
	}

	if err := request.Receive(conn); err != nil {
		log.Printf("Erro ao remover %s do sistema de arquivos. Problema ao receber respostas\n", key)
		return err
	}

	return nil
}

func Get() {

}
