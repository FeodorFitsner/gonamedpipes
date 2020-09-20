package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/pglet/npipe"
)

const (
	readsize = 64 << 10
)

type pipeClient struct {
	id              string
	pageName        string
	sessionID       string
	commandPipeName string
	eventPipeName   string
	events          chan string
	done            chan bool
}

func newPipeClient(pageName string, sessionID string) (*pipeClient, error) {
	id := "111"
	pipeName := fmt.Sprintf("pglet_pipe_%s", id)

	pc := &pipeClient{
		id:              id,
		pageName:        pageName,
		sessionID:       sessionID,
		commandPipeName: pipeName,
		eventPipeName:   pipeName + ".events",
		events:          make(chan string),
	}

	return pc, nil
}

func (pc *pipeClient) start() error {
	//go pc.commandLoop()
	go pc.eventLoop()

	return nil
}

// func (pc *pipeClient) commandLoop() {
// 	log.Println("Starting command loop...")

// 	defer os.Remove(pc.commandPipeName)

// 	for {
// 		// read next command from pipeline
// 		cmdText := pc.read()

// 		// parse command
// 		command, err := page.ParseCommand(cmdText)
// 		if err != nil {
// 			log.Fatalln(err)
// 		}

// 		log.Printf("Send command: %+v", command)

// 		if command.Name == page.Quit {
// 			pc.close()
// 			return
// 		}

// 		rawResult := pc.hostClient.call(page.PageCommandFromHostAction, &page.PageCommandRequestPayload{
// 			PageName:  pc.pageName,
// 			SessionID: pc.sessionID,
// 			Command:   *command,
// 		})

// 		// parse response
// 		payload := &page.PageCommandResponsePayload{}
// 		err = json.Unmarshal(*rawResult, payload)

// 		if err != nil {
// 			log.Fatalln("Error parsing response from PageCommandFromHostAction:", err)
// 		}

// 		// save command results
// 		result := payload.Result
// 		if payload.Error != "" {
// 			result = fmt.Sprintf("error %s", payload.Error)
// 		}

// 		pc.writeResult(result)
// 	}
// }

// func (pc *pipeClient) read() string {
// 	var bytesRead int
// 	var err error
// 	buf := make([]byte, readsize)
// 	for {
// 		var result []byte
// 		input, err := openFifo(pc.commandPipeName, os.O_RDONLY)
// 		if err != nil {
// 			break
// 		}
// 		for err == nil {
// 			bytesRead, err = input.Read(buf)
// 			result = append(result, buf[0:bytesRead]...)

// 			if err == io.EOF {
// 				break
// 			}

// 			//fmt.Printf("read: %d\n", bytesRead)
// 		}
// 		input.Close()
// 		return string(result)
// 	}
// 	log.Fatal(err)
// 	return ""
// }

// func (pc *pipeClient) writeResult(result string) {
// 	log.Println("Waiting for result to consume...")
// 	output, err := openFifo(pc.commandPipeName, os.O_WRONLY)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Println("Write result:", result)

// 	output.WriteString(fmt.Sprintf("%s\n", result))
// 	output.Close()
// }

func (pc *pipeClient) emitEvent(evt string) {
	select {
	case pc.events <- evt:
		// Event sent to queue
	default:
		// No event listeners
	}
}

func (pc *pipeClient) eventLoop() {

	log.Println("Starting event loop - ", pc.eventPipeName)

	ln, err := npipe.Listen(`\\.\pipe\` + pc.eventPipeName)
	if err != nil {
		// handle error
	}

	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}

		log.Println("Connected to event pipe...")

		go func(conn net.Conn) {

			for {
				select {
				case evt, more := <-pc.events:

					w := bufio.NewWriter(conn)

					log.Println("before event written:", evt)
					_, err = w.WriteString(evt + "\n")
					if err != nil {
						if strings.Contains(err.Error(), "Pipe IO timed out waiting") {
							//log.Println("aaa")
							continue
						}
						log.Println("write error:", err)
						return
					}

					conn.SetWriteDeadline(time.Now().Add(10 * time.Millisecond))
					err = w.Flush()
					if err != nil {
						if strings.Contains(err.Error(), "Pipe IO timed out waiting") {
							//log.Println("bbb")
							continue
						}
						log.Println("flush error:", err)
						return
					}

					log.Println("event written:", evt)

					if !more {
						return
					}
				}
			}

		}(conn)

		//log.Println("Disconnected")
	}
}

func (pc *pipeClient) close() {
	log.Println("Closing pipe client...")

	//pc.hostClient.unregisterPipeClient(pc)
}
