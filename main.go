package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

type Twitch struct {
	username       string
	token          string
	channel        string
	conn           net.Conn
	cSend          chan string
	cEvents        chan interface{}
	eventFunctions eventFunctions
}

type eventFunctions struct {
	onConnect []func(*Twitch)
	onPing    []func(*Twitch)
	onMessage []func(string, *Twitch)
}

type eventPinged struct {
	message string
}

type eventConnected struct {
	message string
}

type eventMessageReceived struct {
	message string
}

type connectionError struct {
	err error
}

func (t *Twitch) init() {
	t.cSend = make(chan string)
	t.cEvents = make(chan interface{})
}

func (t *Twitch) Connect() {
	t.init()

	var err error
	t.conn, err = net.Dial("tcp", "irc.chat.twitch.tv:6667")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	go t.send()
	go t.receive()

	t.SendMessage("PASS " + t.token)
	t.SendMessage("NICK " + t.username)

	t.eventTrigger()
}

func (t *Twitch) send() {
	for line := range t.cSend {
		t.conn.Write([]byte(line + "\r\n"))
	}
}

func (t *Twitch) SendMessage(message string) {
	t.cSend <- message
}

func (t *Twitch) receive() {
	buf := bufio.NewReader(t.conn)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			t.cEvents <- connectionError{err: err}
			return
		}

		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "PING"):
			t.cEvents <- eventPinged{message: line}
		case strings.HasPrefix(line, ":tmi.twitch.tv 001"):
			t.cEvents <- eventConnected{message: line}
		default:
			t.cEvents <- eventMessageReceived{message: line}
		}
	}
}

func (t *Twitch) eventTrigger() {
	for event := range t.cEvents {
		switch ev := event.(type) {
		case eventConnected:
			t.SendMessage("JOIN " + t.channel)
			t.SendMessage("PRIVMSG " + t.channel + " :HeyGuys I'm here")
			for _, f := range t.eventFunctions.onConnect {
				go f(t)
			}
		case eventPinged:
			t.SendMessage("PONG :tmi.twitch.tv")
			for _, f := range t.eventFunctions.onPing {
				go f(t)
			}
		case eventMessageReceived:
			for _, f := range t.eventFunctions.onMessage {
				go f(ev.message, t)
			}
		case connectionError:
			fmt.Println(ev.err)
			if ev.err == io.EOF {
				close(t.cEvents)
				close(t.cSend)
				t.Connect()
			} else {
				log.Fatal(ev.err)
			}
		}
	}
}

func (t *Twitch) OnConnect(f func(*Twitch)) {
	t.eventFunctions.onConnect = append(t.eventFunctions.onConnect, f)
}

func (t *Twitch) OnMessage(f func(string, *Twitch)) {
	t.eventFunctions.onMessage = append(t.eventFunctions.onMessage, f)
}

func bla(message string, t *Twitch) {
	fmt.Println(message)
}

func main() {
	token := os.Getenv("TOKEN")

	if token == "" {
		log.Fatal("Empty TOKEN")
	}

	t := Twitch{
		username: "curi0sDE_BOT",
		token:    token,
		channel:  "#curi0sde",
	}

	t.OnMessage(bla)

	t.Connect()
}
