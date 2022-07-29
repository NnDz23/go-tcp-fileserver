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
	SERVER_PORT     = ":8021"
	SERVER_PROTOCOL = "tcp"

	//request format
	//command channel data
	REQUEST         = "%s %s %s\n"
	RESPONSE_PREFIX = "file "
	RESPONSE_REGEX  = RESPONSE_PREFIX + "{.+}\n"

	FILES_DIR = "./files/"
)

var wg sync.WaitGroup
var allowOverwrite bool

type fileContent struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Content   string `json:"content"`
}

func HandleSubscribe(subscribeCmd *flag.FlagSet, channel *string) {
	//Parse args
	subscribeCmd.Parse(os.Args[2:])
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

// Reads files from the connection.
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
		if match {
			fileContentStr := strings.ReplaceAll(str, RESPONSE_PREFIX, "")
			fileContentStr = strings.ReplaceAll(fileContentStr, "\n", "")
			var file fileContent
			err = json.Unmarshal([]byte(fileContentStr), &file)
			if err != nil {
				log.Println(err)
				continue
			}
			err = SaveFile(file, c)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

func HandleSend(sendCmd *flag.FlagSet, file *string, channel *string) {
	//Parse args
	sendCmd.Parse(os.Args[2:])
	//Channel validation
	if *channel == "" {
		log.Println("expected a channel to receive files")
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

func ValidateFile(file *string) error {
	_, err := os.Stat(*file)
	return err
}

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

func SaveFile(f fileContent, c string) error {
	dec, err := base64.StdEncoding.DecodeString(f.Content)
	if err != nil {
		log.Println(err)
	}
	//create channel directory if it does not exists
	path := FILES_DIR + c + "/"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	filePath := path + f.Name + f.Extension
	var writeFile bool
	if _, err := os.Stat(filePath); err == nil {
		if allowOverwrite {
			log.Println("overwriting", filePath)
		}
		writeFile = allowOverwrite
	} else {
		writeFile = true
	}

	if writeFile {
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
	} else {
		log.Println(filePath, "already exists, overwriting not allowed")
	}

	return nil
}

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
