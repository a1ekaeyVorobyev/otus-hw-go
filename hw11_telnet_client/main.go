package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Place your code here
	// P.S. Do not rush to throw context down, think think if it is useful with blocking operation?
	var timeout time.Duration
	if len(os.Args) < 3 {
		fmt.Printf("Set address server and port: example 127.0.0.1  25 --timeout=10s , default timeout = 10s\n")
		os.Exit(1)
	}
	flag.DurationVar(&timeout, "timeout", time.Second*10, "time out in seconds")
	flag.Parse()
	arg := flag.Args()
	host := arg[0]
	port := arg[1]
	c := NewTelnetClient(net.JoinHostPort(host, port), timeout, os.Stdin, os.Stdout)

	err := c.Connect()
	if err != nil {
		fmt.Printf("Can't connect to telnet server: %s. \nError -  %s\n", c.GetAdress(), err.Error())
		os.Exit(1)
	}

	ch := make(chan bool, 1)

	go hanlerRead(c, ch)
	go handlerwriter(c, ch)

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
		<-sigs
		ch <- true
	}()

exit:
	for {
		select {
		case <-ch:
			break exit
		}
	}
	log.Println("Exit telnet.")
}

func hanlerRead(tc TelnetClient, ch chan bool) {
	defer func() { ch <- true }()

	if err := tc.Receive(); err != nil { // if server close connect this routine is exit but we wait some unsuccessful attempts to send in writeRoutine
		fmt.Fprintf(os.Stderr, "%v\n", err)

		return
	}
}

func handlerwriter(tc TelnetClient, ch chan bool) {
	defer func() { ch <- true }()

	if err := tc.Send(); err != nil {
		fmt.Fprintf(os.Stderr, "...Connection was closed by peer\n") // an error occurs if server sent ctrl + c (close) and client execute some unsuccessful attempts to send

		return
	}
}
