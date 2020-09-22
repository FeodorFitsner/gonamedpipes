// +build windows

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/pglet/npipe"
)

const (
	readsize = 64 << 10
)

type pipeImpl struct {
	conn            net.Conn
	id              string
	commandPipeName string
	eventPipeName   string
	commands        chan string
	events          chan string
	done            chan bool
}

func newPipeImpl(id string) (*pipeImpl, error) {
	pipeName := fmt.Sprintf("pglet_pipe_%s", id)

	pc := &pipeImpl{
		id:              id,
		commandPipeName: pipeName,
		eventPipeName:   pipeName + ".events",
		commands:        make(chan string),
		events:          make(chan string),
	}

	go pc.commandLoop()
	go pc.eventLoop()

	return pc, nil
}

func (pc *pipeImpl) commandLoop() {
	log.Println("Starting command loop - ", pc.commandPipeName)

	ln, err := npipe.Listen(`\\.\pipe\` + pc.commandPipeName)
	if err != nil {
		// handle error
	}

	defer ln.Close()

	for {
		pc.conn, err = ln.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}

		log.Println("Connected to command pipe...")

		go func() {

			for {
				// read next command from pipeline
				cmdText := pc.read()

				if cmdText == "" {
					log.Println("Disconnected from command pipe")
					return
				}

				pc.commands <- cmdText
			}

		}()
	}
}

func (pc *pipeImpl) read() string {

	var bytesRead int
	var err error
	buf := make([]byte, readsize)

	r := bufio.NewReader(pc.conn)

	log.Println("Before read")

	for {
		var result []byte

		for {

			bytesRead, err = r.Read(buf)

			if err == io.EOF {
				//log.Println("EOF")
				return ""
			}

			result = append(result, buf[0:bytesRead]...)

			//log.Println(string(result))

			//log.Printf("read: %d\n", bytesRead)

			if bytesRead < readsize {
				//log.Println("less bytes read")
				break
			}
		}
		return strings.TrimSuffix(strings.TrimSuffix(string(result), "\n"), "\r")
	}
}

func (pc *pipeImpl) writeResult(result string) {
	log.Println("Waiting for result to consume...")

	w := bufio.NewWriter(pc.conn)

	log.Println("Write result:", result)

	w.WriteString(fmt.Sprintf("%s\n", result))
	w.Flush()
}

func (pc *pipeImpl) emitEvent(evt string) {
	select {
	case pc.events <- evt:
		// Event sent to queue
	default:
		// No event listeners
	}
}

func (pc *pipeImpl) eventLoop() {

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

			defer log.Println("Disconnected from event pipe")

			for {
				select {
				case evt, more := <-pc.events:

					w := bufio.NewWriter(conn)

					//log.Println("before event written:", evt)
					_, err = w.WriteString(evt + "\n")
					if err != nil {
						if strings.Contains(err.Error(), "Pipe IO timed out waiting") {
							continue
						}
						//log.Println("write error:", err)
						return
					}

					conn.SetWriteDeadline(time.Now().Add(10 * time.Millisecond))
					err = w.Flush()
					if err != nil {
						if strings.Contains(err.Error(), "Pipe IO timed out waiting") {
							continue
						}
						//log.Println("flush error:", err)
						return
					}

					log.Println("event written:", evt)

					if !more {
						return
					}
				}
			}

		}(conn)
	}
}

func (pc *pipeImpl) close() {
	log.Println("Closing pipe client...")

	//pc.hostClient.unregisterPipeClient(pc)
}
