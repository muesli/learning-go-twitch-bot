package bot

const (
	TypeConnected = iota
	TypePinged
	TypeMessageReceived
)

type EventType int

type EventPinged struct {
	Message string
}

type EventConnected struct {
	Message string
}

type EventMessageReceived struct {
	Message string
	User    string
}

type ConnectionError struct {
	Err error
}

func (t *Twitch) eventTrigger() chan interface{} {
	ch := make(chan interface{})

	go func() {
		for event := range t.cEvents {
			ch <- event

			switch event.(type) {
			case EventConnected:
				t.SendMessage("JOIN " + t.opts.Channel)

			case EventPinged:
				t.SendMessage("PONG :tmi.twitch.tv")
			}
		}
	}()

	return ch
}
