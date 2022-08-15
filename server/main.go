package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const (
	// SERVER_API_PORT port to serve API
	SERVER_API_PORT = 8081
	//SERVER_PORT port for fileserver protocol
	SERVER_PORT = 8021
	//SERVER_PROTOCOL protocol to use for fileserver
	SERVER_PROTOCOL = "tcp"

	//SUBSCRIBE_REQUEST_PREFIX subscribe command prefix
	SUBSCRIBE_REQUEST_PREFIX = "subscribe "
	SUBSCRIBE_REQUEST_REGEX  = SUBSCRIBE_REQUEST_PREFIX + ".+ {}"
	SUBSCRIBE_RESPONSE       = "file %s\n"

	SEND_REQUEST_PREFIX = "send "
	SEND_REQUEST_REGEX  = SEND_REQUEST_PREFIX + ".+ {.+}"
	SEND_REQUEST        = "%s %s %s\n"
)

var (
	errUnexistingChannel = errors.New("there was an attempt to send a file to an unexisting channel")
)

type fileContent struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Content   string `json:"content"`
}

type Channel struct {
	Name              string              `json:"name"`
	ClientConnections map[string]net.Conn `json:"-"`
	Files             int                 `json:"files_sent"`
	Clients           int                 `json:"clients_connected"`
	CreatedAt         time.Time           `json:"created_at"`
}

type ServerStats struct {
	Files     int       `json:"files_sent"`
	Clients   int       `json:"clients_connected"`
	Channels  int       `json:"channels_available"`
	CreatedAt time.Time `json:"created_at"`
}

type Server struct {
	channels    map[string]*Channel
	sendRequest chan string
	started_at  time.Time
}

func NewServer() *Server {
	server := &Server{
		channels:    make(map[string]*Channel),
		sendRequest: make(chan string),
		started_at:  time.Now(),
	}
	server.Listen()
	return server
}

func NewChannel(client net.Conn, name string) *Channel {
	channel := &Channel{
		Name:              name,
		ClientConnections: map[string]net.Conn{client.RemoteAddr().String(): client},
		Files:             0,
		Clients:           1,
		CreatedAt:         time.Now(),
	}
	return channel
}

func (server *Server) GetServerStats() *ServerStats {
	files := 0
	clients := 0
	channels := len(server.channels)

	for _, channel := range server.channels {
		files += channel.Files
		clients += channel.Clients
	}

	serverStats := &ServerStats{
		Files:     files,
		Clients:   clients,
		Channels:  channels,
		CreatedAt: server.started_at,
	}

	return serverStats
}

func (server *Server) SubscribeClient(client net.Conn, channel string) {
	if serverChannel, ok := server.channels[channel]; ok {
		serverChannel.Clients = serverChannel.Clients + 1
		serverChannel.ClientConnections[client.RemoteAddr().String()] = client
	} else {
		c := NewChannel(client, channel)
		server.channels[channel] = c
	}
}

func (server *Server) UnsubscribeClient(client net.Conn, channel string) {
	clientKey := client.RemoteAddr().String()
	serverChannel := server.channels[channel]
	delete(serverChannel.ClientConnections, clientKey)
	serverChannel.Clients = serverChannel.Clients - 1
	err := client.Close()
	if err != nil {
		log.Println(err)
	}
}

func (server *Server) Listen() {
	go func() {
		for {
			request := <-server.sendRequest
			parsedFileContent, channel, err := server.Parse(request)

			if err != nil {
				log.Printf("listen error: %s", err.Error())
				continue
			}

			server.Broadcast(channel, parsedFileContent)
		}
	}()
}

// Parse converts request ands returns parsedFileContent and channel
func (server *Server) Parse(request string) (string, string, error) {
	sendRequest := strings.ReplaceAll(request, SEND_REQUEST_PREFIX, "")
	s := strings.Split(sendRequest, " ")
	channel, fileContentStr := s[0], strings.Join(s[1:], " ")
	if _, ok := server.channels[channel]; !ok {
		return "", "", errUnexistingChannel
	}

	var file fileContent
	err := json.Unmarshal([]byte(fileContentStr), &file)
	if err != nil {
		return "", "", err
	}

	return fileContentStr, channel, nil
}

// Broadcast sends fileContentStr to all clients in channel
func (server *Server) Broadcast(channel string, fileContentStr string) {
	serverChannel := server.channels[channel]
	serverChannel.Files = serverChannel.Files + 1

	for client, conn := range server.channels[channel].ClientConnections {
		log.Println("sending file to", client)
		_, err := conn.Write([]byte(fmt.Sprintf(SUBSCRIBE_RESPONSE, fileContentStr)))
		if err != nil {
			log.Println(err)
			//unsubscribe client when there is an error
			log.Println("unsubscribing", client)
			server.UnsubscribeClient(conn, channel)
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
	match, err := regexp.MatchString(SEND_REQUEST_REGEX, strings.TrimSuffix(req, "\n"))
	if err != nil {
		log.Println("Error when checking for match")
		match = false
	}
	//Send request
	if match {
		server.sendRequest <- strings.TrimSuffix(req, "\n")
	}
	match, err = regexp.MatchString(SUBSCRIBE_REQUEST_REGEX, strings.TrimSuffix(req, "\n"))
	if err != nil {
		log.Println("Error when checking for match")
		match = false
	}
	//Subscribe request
	if match {
		subscribeRequest := strings.ReplaceAll(req, SUBSCRIBE_REQUEST_PREFIX, "")
		s := strings.Split(subscribeRequest, " ")
		channel := s[0]
		server.SubscribeClient(conn, channel)
	}

}

func WriteJsonResponse(w http.ResponseWriter, status int, output []byte) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err := w.Write(output)
	if err != nil {
		return err
	}

	return nil
}

func (server *Server) ServeAPI() {
	log.Println("API listening on port", SERVER_API_PORT)

	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	mux.Get("/channels/list", func(w http.ResponseWriter, r *http.Request) {

		// Convert map to slice of values.
		channels := []Channel{}
		for _, value := range server.channels {
			channels = append(channels, *value)
		}

		output, err := json.Marshal(channels)
		if err != nil {
			log.Println(err)
		}
		err = WriteJsonResponse(w, http.StatusOK, output)
		if err != nil {
			log.Println(err)
		}
	})

	mux.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		stats := *server.GetServerStats()
		output, err := json.Marshal(stats)
		if err != nil {
			log.Println(err)
		}
		err = WriteJsonResponse(w, http.StatusOK, output)
		if err != nil {
			log.Println(err)
		}

	})

	mux.Post("/channels/send", func(w http.ResponseWriter, r *http.Request) {
		type body struct {
			Channel   string `json:"channel"`
			Name      string `json:"name"`
			Extension string `json:"extension"`
			Base64    string `json:"base64"`
		}
		type jsonResponse struct {
			Error   bool   `json:"error"`
			Message string `json:"message"`
		}

		var b body
		var res jsonResponse

		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&b)

		if err != nil {
			//write error response
			res.Error = true
			res.Message = "error on content received"
			output, err := json.Marshal(res)
			if err != nil {
				log.Println(err)
			}
			err = WriteJsonResponse(w, http.StatusBadRequest, output)
			if err != nil {
				log.Println(err)
			}
			return
		}
		err = dec.Decode(&struct{}{}) //check for single json
		if err != io.EOF {
			//write error response
			res.Error = true
			res.Message = "error on content received"
			output, err := json.Marshal(res)
			if err != nil {
				log.Println(err)
			}
			err = WriteJsonResponse(w, http.StatusBadRequest, output)
			if err != nil {
				log.Println(err)
			}
			return
		}

		var f fileContent
		f.Name = b.Name
		f.Extension = b.Extension
		f.Content = b.Base64

		fileToSendJson, err := json.Marshal(f)
		if err != nil {
			//write error response
			res.Error = true
			res.Message = "error while sending file"
			output, err := json.Marshal(res)
			if err != nil {
				log.Println(err)
			}
			err = WriteJsonResponse(w, http.StatusBadRequest, output)
			if err != nil {
				log.Println(err)
			}
			return
		}
		request := fmt.Sprintf(SEND_REQUEST, "send", b.Channel, string(fileToSendJson))
		server.sendRequest <- strings.TrimSuffix(request, "\n")
		res.Error = false
		res.Message = "file sent succesfully"
		output, err := json.Marshal(res)
		if err != nil {
			log.Println(err)
		}
		err = WriteJsonResponse(w, http.StatusOK, output)
		if err != nil {
			log.Println(err)
		}
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", SERVER_API_PORT),
		Handler: mux,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Println("Error when starting API")
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

		//Start API server
		go server.ServeAPI()

		//Start fileserver
		listener, err := net.Listen(SERVER_PROTOCOL, fmt.Sprintf(":%d", SERVER_PORT))
		if err != nil {
			log.Println("Error: ", err)
			os.Exit(1)
		}
		defer listener.Close()
		log.Println("Listening on " + fmt.Sprintf(":%d", SERVER_PORT))

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
