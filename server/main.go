package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

const (
	SERVER_PORT     = ":8021"
	SERVER_PROTOCOL = "tcp"

	SUBSCRIBE_REQUEST_PREFIX = "subscribe "
	SUBSCRIBE_REQUEST_REGEX  = SUBSCRIBE_REQUEST_PREFIX + ".+ {}"
	SUBSCRIBE_RESPONSE       = "file %s\n"

	SEND_REQUEST_PREFIX = "send "
	SEND_REQUEST_REGEX  = SEND_REQUEST_PREFIX + ".+ {.+}"
)

type fileContent struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Content   string `json:"content"`
}

type Server struct {
	channels    map[string][]net.Conn
	sendRequest chan string
}

func NewServer() *Server {
	server := &Server{
		channels:    make(map[string][]net.Conn),
		sendRequest: make(chan string),
	}
	server.Listen()
	return server
}

func (server *Server) Listen() {
	go func() {
		for {
			request := <-server.sendRequest
			server.Parse(request)
		}
	}()
}

func (server *Server) Parse(request string) {
	sendRequest := strings.ReplaceAll(request, SEND_REQUEST_PREFIX, "")
	s := strings.Split(sendRequest, " ")
	channel, fileContentStr := s[0], strings.Join(s[1:], " ")
	var file fileContent
	err := json.Unmarshal([]byte(fileContentStr), &file)
	if err != nil {
		log.Println(err)
		return
	}
	server.Broadcast(channel, fileContentStr)
}

func (server *Server) Broadcast(channel string, fileContestStr string) {
	for connIdx := range server.channels[channel] {
		conn := server.channels[channel][connIdx]
		writer := bufio.NewWriter(conn)
		_, err := writer.Write([]byte(fmt.Sprintf(SUBSCRIBE_RESPONSE, fileContestStr)))
		if err != nil {
			log.Println(err)
		}
	}
}

func (server *Server) HandleConnection(conn net.Conn) {
	//Check if its subscribe or send
	reader := bufio.NewReader(conn)
	req, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
		return
	}
	match, _ := regexp.MatchString(SEND_REQUEST_REGEX, strings.TrimSuffix(req, "\n"))
	//Send request
	if match {
		server.sendRequest <- strings.TrimSuffix(req, "\n")
	}
	match, _ = regexp.MatchString(SUBSCRIBE_REQUEST_REGEX, strings.TrimSuffix(req, "\n"))
	//Subscribe request
	if match {
		subscribeRequest := strings.ReplaceAll(req, SUBSCRIBE_REQUEST_PREFIX, "")
		s := strings.Split(subscribeRequest, " ")
		channel := s[0]
		server.channels[channel] = append(server.channels[channel], conn)
	}

}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) < 2 {
		fmt.Println("expected at least one command")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "start": // if its the start command
		server := NewServer()

		listener, err := net.Listen(SERVER_PROTOCOL, SERVER_PORT)
		if err != nil {
			log.Println("Error: ", err)
			os.Exit(1)
		}
		defer listener.Close()
		log.Println("Listening on " + SERVER_PORT)

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Error: ", err)
				continue
			}
			server.HandleConnection(conn)
		}
	default: // if we don't understand the input
		log.Println("unkown command")
	}

}
