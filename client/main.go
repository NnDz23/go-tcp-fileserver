package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/manifoldco/promptui"
)

const (
	// SERVER_PORT port for fileserver protocol
	SERVER_PORT = ":8021"
	// SERVER_PROTOCOL protocol to use for fileserver
	SERVER_PROTOCOL = "tcp"

	//request format
	//command channel data

	// REQUEST format of the request
	REQUEST = "%s %s %s\n"
	// RESPONSE_PREFIX preffix for a received file
	RESPONSE_PREFIX = "file "
	// RESPONSE_REGEX regular expression to check if incoming file complies
	RESPONSE_REGEX = RESPONSE_PREFIX + "{.+}\n"
	// FILES_DIR directory used to store incoming files
	FILES_DIR = "./files/"
)

var (
	// errUnexistingChannel error for unexisting channels
	errOverwriteNotAllowed = errors.New("overwrite required, but not allowed")
)

var wg sync.WaitGroup
var allowOverwrite bool

// fileContent struct that contains metadata (Name and Extension) associated to the Content
type fileContent struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Content   string `json:"content"`
}

// HandleSubscribe subscribes client to start receiving files from channel
func HandleSubscribe(subscribeCmd *flag.FlagSet, channel *string) {
	//Parse args
	err := subscribeCmd.Parse(os.Args[2:])
	if err != nil {
		log.Println(err)
		return
	}
	//Channel validation
	if *channel == "" {
		log.Println("expected a channel to receive files")
		return
	}
	wg.Add(1)

	//Preparing request
	request := fmt.Sprintf(REQUEST, "subscribe", *channel, "{}")

	conn, err := net.Dial(SERVER_PROTOCOL, SERVER_PORT)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()
	//Sending request to server
	writer := bufio.NewWriter(conn)
	_, err = writer.WriteString(request)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = writer.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	//Reading for incoming messages on channel
	go Read(conn, *channel)

	wg.Wait()
}

// Read reads for incomming files on the connection for channel c
func Read(conn net.Conn, c string) {
	reader := bufio.NewReader(conn)
	r, _ := regexp.Compile(RESPONSE_REGEX)
	log.Println("Allow incoming files to overwrite existing ones?")
	allowOverwrite = cmdYesNo()
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			wg.Done()
			return
		}
		match := r.MatchString(str)
		if !match {
			log.Println("received message is not an appropiate file")
			continue
		}
		fileContentStr := strings.ReplaceAll(str, RESPONSE_PREFIX, "")
		fileContentStr = strings.ReplaceAll(fileContentStr, "\n", "")
		var file fileContent
		err = json.Unmarshal([]byte(fileContentStr), &file)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("received file", file.Name, "from", c)
		err = SaveFile(file, c)
		if err != nil {
			log.Println(err)
		}
	}
}

// HandleSend sends the file from the client to the channel
func HandleSend(sendCmd *flag.FlagSet, file *string, channel *string) {
	//Parse args
	err := sendCmd.Parse(os.Args[2:])
	if err != nil {
		log.Println(err)
		return
	}
	//Channel validation
	if *channel == "" {
		log.Println("expected a channel to send files to")
		return
	}
	//File validations
	if *file == "" {
		log.Println("expected a file to send")
		return
	}
	if err := ValidateFile(file); err != nil {
		log.Println("invalid file")
		return
	}
	//Getting file
	fileToSend, err := GetFile(file)
	if err != nil {
		log.Println("error when getting file")
		return
	}
	//Marshalling file
	fileToSendJson, err := json.Marshal(fileToSend)
	if err != nil {
		log.Println(err)
		return
	}
	//Formatting request
	request := fmt.Sprintf(REQUEST, "send", *channel, string(fileToSendJson))
	//Connecting to server
	conn, err := net.Dial(SERVER_PROTOCOL, SERVER_PORT)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()
	//Sending request to server
	writer := bufio.NewWriter(conn)
	_, err = writer.WriteString(request)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = writer.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

}

// ValidateFile validates if specified file is an actual file
func ValidateFile(file *string) error {
	_, err := os.Stat(*file)
	return err
}

// GetFile gets fileContent for the specified file
func GetFile(file *string) (*fileContent, error) {
	//get file extension
	extension := filepath.Ext(*file)

	//get file name
	name := strings.TrimSuffix(filepath.Base(*file), filepath.Ext(*file))

	//get data as base64
	//open file
	f, err := os.Open(*file)
	if err != nil {
		return nil, err
	}
	//read into byte slice
	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	//encode as base64
	data := base64.StdEncoding.EncodeToString(content)

	fc := &fileContent{
		Name:      name,
		Extension: extension,
		Content:   data,
	}

	return fc, nil
}

// SaveFile saves received file to the files directory
func SaveFile(f fileContent, c string) error {
	dec, err := base64.StdEncoding.DecodeString(f.Content)
	if err != nil {
		return err
	}
	//create channel directory if it does not exists
	path := FILES_DIR + c + "/"
	var pathExists bool
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		pathExists = false
	} else if err != nil {
		pathExists = true
	} else {
		return err
	}

	if !pathExists {
		err = os.Mkdir(path, os.ModePerm)
	}
	if err != nil {
		return err
	}

	filePath := path + f.Name + f.Extension
	overwriteRequired := false
	// if error is returned then there's no existing file with that name
	// this is without checking if we have required permissions
	if _, err := os.Stat(filePath); err == nil {
		overwriteRequired = true
	}

	if overwriteRequired && !allowOverwrite {
		return errOverwriteNotAllowed
	}

	//create file
	file, err := os.Create(path + f.Name + f.Extension)
	if err != nil {
		return err
	}
	defer file.Close()
	//write file
	if _, err := file.Write(dec); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}

	return nil
}

// cmdYesNo prompts user for a Yes/No, returns true if Yes
func cmdYesNo() bool {
	prompt := promptui.Select{
		Label: "Select [Yes/No]",
		Items: []string{"Yes", "No"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}
	return result == "Yes"
}

func main() {
	//subscribe command
	subscribeCmd := flag.NewFlagSet("subscribe", flag.ExitOnError)
	subscribeChannel := subscribeCmd.String("c", "", "name of the channel you want to subscribe to")

	//send command
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendFile := sendCmd.String("f", "", "path of the file you want to send")
	sendChannel := sendCmd.String("c", "", "name of the channel you want to send the file to")

	if len(os.Args) < 2 {
		log.Println("expected at least one command")
		os.Exit(1)
	}
	switch os.Args[1] {
	case subscribeCmd.Name():
		HandleSubscribe(subscribeCmd, subscribeChannel)
	case sendCmd.Name():
		HandleSend(sendCmd, sendFile, sendChannel)
	default:
		log.Println("unkown command", os.Args[1])
	}
}
