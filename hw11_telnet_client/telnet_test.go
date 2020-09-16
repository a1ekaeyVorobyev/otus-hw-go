package main

import (
	"bytes"
	"io/ioutil"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"syreclabs.com/go/faker"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, ioutil.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})
}

func TestTelnetMy(t *testing.T) {
	var wg sync.WaitGroup
	text := faker.Company().CatchPhrase() + "\n"
	timeout := time.Duration(time.Second * 10)
	var inMessage, result string
	ch := make(chan string)
	chCheck := make(chan bool)

	host := "127.0.0.1:1235"
	wg.Add(4)
	t.Run("Check Connect Client", func(t *testing.T) {
		out := &bytes.Buffer{}
		in := &bytes.Buffer{}
		go func(host string, ch chan string, chCheck chan bool) {
			defer wg.Done()
			l, err := net.Listen("tcp", host)
			require.NoError(t, err)
			defer func() { require.NoError(t, l.Close()) }()
			chCheck <- true
			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()
			ch <- conn.LocalAddr().String()

			n, err := conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}(host, ch, chCheck)
		<-chCheck
		c := NewTelnetClient(host, timeout, ioutil.NopCloser(in), out)
		err := c.Connect()
		require.NoError(t, err)
		inMessage = <-ch
		result = c.GetAdress()
		require.Equal(t, result, inMessage)
		err = c.Receive()
		require.NoError(t, err)
		require.Equal(t, "world\n", out.String())
		c.Close()
	})
	t.Run("Check Connect Recive", func(t *testing.T) {
		out := &bytes.Buffer{}
		in := &bytes.Buffer{}
		go func(host, text string, chCheck chan bool) {
			defer wg.Done()
			l, err := net.Listen("tcp", host)
			require.NoError(t, err)
			defer func() { require.NoError(t, l.Close()) }()
			chCheck <- true
			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()
			n, err := conn.Write([]byte(text))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}(host, text, chCheck)
		<-chCheck
		c := NewTelnetClient(host, timeout, ioutil.NopCloser(in), out)
		err := c.Connect()
		require.NoError(t, err)
		result = c.GetAdress()
		require.Equal(t, result, inMessage)
		err = c.Receive()
		require.NoError(t, err)
		require.Equal(t, text, out.String())
		c.Close()
	})
	t.Run("Check Connect Send", func(t *testing.T) {
		out := &bytes.Buffer{}
		in := &bytes.Buffer{}
		go func(host, text string, chCheck chan bool) {
			defer wg.Done()
			l, err := net.Listen("tcp", host)
			require.NoError(t, err)
			defer func() { require.NoError(t, l.Close()) }()
			chCheck <- true
			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()
			request := make([]byte, 1024)
			n, err := conn.Read([]byte(request))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}(host, text, chCheck)
		<-chCheck
		c := NewTelnetClient(host, timeout, ioutil.NopCloser(in), out)
		err := c.Connect()
		require.NoError(t, err)
		in.WriteString(text)
		err = c.Send()
		require.NoError(t, err)
		c.Close()
	})
	t.Run("Check Connect Send and Recive", func(t *testing.T) {
		out := &bytes.Buffer{}
		in := &bytes.Buffer{}
		go func(host, text string, chCheck chan bool) {
			defer wg.Done()
			l, err := net.Listen("tcp", host)
			require.NoError(t, err)
			defer func() { require.NoError(t, l.Close()) }()
			chCheck <- true
			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()
			request := make([]byte, 1024)
			n, err := conn.Read([]byte(request))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
			n, err = conn.Write([]byte(request)[:n])
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}(host, text, chCheck)
		<-chCheck
		c := NewTelnetClient(host, timeout, ioutil.NopCloser(in), out)
		err := c.Connect()
		require.NoError(t, err)
		in.WriteString(text)
		err = c.Send()
		require.NoError(t, err)
		err = c.Receive()
		require.NoError(t, err)
		require.Equal(t, text, out.String())
		c.Close()
	})
	wg.Wait()
}
