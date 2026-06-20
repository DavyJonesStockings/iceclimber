package main

type Event struct {
	// fields TBD once you settle on what Neovim sends
}

func listenIPC(onEvent func(Event)) {
	// open the Unix socket, read incoming events, call onEvent(parsedEvent) per message
}
