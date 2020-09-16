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
	stimeout := ""
	if len(os.Args) < 3 {
		fmt.Printf("Set address server and port: example 127.0.0.1  25 --timeout=10s , default timeout = 10s\n")
		os.Exit(1)
	}
	flag.StringVar(&stimeout, "timeout", "10s", "time out in seconds")
	flag.Parse()
	host := os.Args[1]
	port := os.Args[2]
	if len(os.Args) > 3 {
		host = os.Args[2]
		port = os.Args[3]
	}
	timeout, err := time.ParseDuration(stimeout)
	if err != nil {
		log.Fatal("Please enter correct timeout value", err.Error())
	}
	c := NewTelnetClient(net.JoinHostPort(host, port), timeout, os.Stdin, os.Stdout)
	ctx := c.GetContext()
	err = c.Connect()
	if err != nil {
		fmt.Printf("Can't connect to telnet server: %s. \nError -  %s\n", c.GetAdress(), err.Error())
		os.Exit(1)
	}
	ch := make(chan bool, 1)
	go hanlerRead(c)
	go handlerwriter(c)
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
		<-sigs
		c.Cancel()
	}()

exit:
	for {
		select {
		case <-ctx.Done():
			break exit
		case <-ch:
			break exit
		}
	}
	log.Println("Exit telnet.")
}

func hanlerRead(tc TelnetClient) {
	if err := tc.Receive(); err != nil { // if server close connect this routine is exit but we wait some unsuccessful attempts to send in writeRoutine
		fmt.Fprintf(os.Stderr, "%v\n", err)
		tc.Cancel()

		return
	}
}

func handlerwriter(tc TelnetClient) {
	if err := tc.Send(); err != nil {
		fmt.Fprintf(os.Stderr, "...Connection was closed by peer\n") // an error occurs if server sent ctrl + c (close) and client execute some unsuccessful attempts to send
		tc.Cancel()
	}
}
