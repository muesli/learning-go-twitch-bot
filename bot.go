package bot

import (
	"bufio"
	"log"
	"net"
	"os"
	"strings"
)

// Options are used to configure the bot.
type Options struct {
	Username string
	Token    string
	Channel  string
}

// Twitch is a twitch IRC bot.
type Twitch struct {
	opts Options

	conn    net.Conn
	cSend   chan string
	cEvents chan interface{}
}

// NewTwitch returns a new Twitch bot.
func NewTwitch(options Options) *Twitch {
	return &Twitch{
		opts:    options,
		cSend:   make(chan string),
		cEvents: make(chan interface{}),
	}
}

func (t *Twitch) Options() Options {
	return t.opts
}

// Connect establishes a connection to the IRC server,
// returning an event channel.
func (t *Twitch) Connect() chan interface{} {
	var err error
	t.conn, err = net.Dial("tcp", "irc.chat.twitch.tv:6667")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	go t.send()
	go t.receive()

	t.SendMessage("PASS " + t.opts.Token)
	t.SendMessage("NICK " + t.opts.Username)

	return t.eventTrigger()
}

func (t *Twitch) send() {
	for line := range t.cSend {
		t.conn.Write([]byte(line + "\r\n"))
	}
}

// SendMessage sends a message
func (t *Twitch) SendMessage(message string) {
	t.cSend <- message
}

func (t *Twitch) receive() {
	buf := bufio.NewReader(t.conn)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			t.cEvents <- ConnectionError{Err: err}
			return
		}

		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "PING"):
			t.cEvents <- EventPinged{Message: line}
		case strings.HasPrefix(line, ":tmi.twitch.tv 001"):
			t.cEvents <- EventConnected{Message: line}
		default:
			t.cEvents <- EventMessageReceived{Message: line, User: "fabian"}
		}
	}
}
