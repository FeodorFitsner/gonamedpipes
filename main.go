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

func main() {
	fmt.Println("Hello, world!")
	//exampleListen()
	pc, _ := newPipeImpl("111")

	go func() {
		for {
			// command echo loop
			cmd := <-pc.commands
			log.Println(cmd)

			pc.writeResult(fmt.Sprintf("%s - OK", strings.TrimSpace(strings.Split(cmd, "\n")[0])))
		}
	}()

	for i := 0; i < 10000; i++ {
		pc.emitEvent(fmt.Sprintf("event %d", i))
		time.Sleep(time.Millisecond * 10)
	}
}

func exampleListen() {
	ln, err := npipe.Listen(`\\.\pipe\mypipe`)
	if err != nil {
		// handle error
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			continue
		}

		// handle connection like any other net.Conn
		go func(conn net.Conn) {
			//r := bufio.NewReader(conn)
			w := bufio.NewWriter(conn)

			// msg, err := r.ReadString('\n')
			// if err != nil {
			// 	// handle error
			// 	return
			// }
			// fmt.Println(strings.TrimSpace(msg))

			time.Sleep(time.Second * 3)

			w.WriteString("Line1\n")
			w.Flush()
			fmt.Println("Line 1 sent")

			time.Sleep(time.Second * 3)

			w.WriteString("Line2\n")
			w.Flush()
			fmt.Println("Line 2 sent")

		}(conn)
	}
}
