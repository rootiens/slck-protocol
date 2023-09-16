package main

import "strings"

type Hub struct {
	Channels        map[string]*Channel
	Clients         map[string]*Client
	Commands        chan Command
	Deregistrations chan *Client
	Registrations   chan *Client
}

func newHub() *Hub {
	return &Hub{
		Registrations:   make(chan *Client),
		Deregistrations: make(chan *Client),
		Clients:         make(map[string]*Client),
		Channels:        make(map[string]*Channel),
		Commands:        make(chan Command),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.Registrations:
			h.Register(client)
		case client := <-h.Deregistrations:
			h.Deregister(client)
		case cmd := <-h.Commands:
			switch cmd.Id {
			case JOIN:
				h.JoinChannel(cmd.Sender, cmd.Recipient)
			case LEAVE:
				h.LeaveChannel(cmd.Sender, cmd.Recipient)
			case MSG:
				h.Message(cmd.Sender, cmd.Recipient, cmd.Body)
			case USRS:
				h.ListUsers(cmd.Sender)
			case CHNS:
				h.ListChannels(cmd.Sender)
			default:
				// Freak out?
			}
		}
	}
}

func (h *Hub) Register(c *Client) {
	if _, exists := h.Clients[c.Username]; exists {
		c.Username = ""
		c.Conn.Write([]byte("ERR username taken\n"))
	} else {
		h.Clients[c.Username] = c
		c.Conn.Write([]byte("OK\n"))
	}
}

func (h *Hub) Deregister(c *Client) {
	if _, exists := h.Clients[c.Username]; exists {
		delete(h.Clients, c.Username)

		for _, channel := range h.Channels {
			delete(channel.Clients, c)
		}
	}
}

func (h *Hub) JoinChannel(u string, c string) {
	// if channel didn't exist
	if _, exists := h.Channels[c]; !exists {
		h.Channels[c] = newChannel(c)
	}

	channel := h.Channels[c]
	client := h.Clients[u]

	channel.Clients[client] = true
}

func (h *Hub) LeaveChannel(u string, c string) {
	if client, ok := h.Clients[u]; ok {
		if channel, ok := h.Channels[c]; ok {
			delete(channel.Clients, client)
		}
	}
}

func (h *Hub) Message(username string, recipient string, message []byte) {
	if sender, ok := h.Clients[username]; ok {
		switch recipient[0] {
		case '#':
			if channel, ok := h.Channels[recipient]; ok {
				if _, ok := channel.Clients[sender]; ok {
					channel.broadcast(sender.Username, message)
				}
			}

		case '@':
			if user, ok := h.Clients[recipient]; ok {
				user.Conn.Write(append(message, '\n'))
			}
		}
	}
}

func (h *Hub) ListChannels(u string) {
	if client, ok := h.Clients[u]; ok {
		var names []string

		if len(h.Channels) == 0 {
			client.Conn.Write([]byte("ERR no channels found\n"))
		}

		for c := range h.Channels {
			names = append(names, "#"+c+" ")
		}

		resp := strings.Join(names, ", ")

		client.Conn.Write([]byte(resp + "\n"))
	}
}

func (h *Hub) ListUsers(u string) {
	if client, ok := h.Clients[u]; ok {
		var names []string

		for c, _ := range h.Clients {
			names = append(names, "@"+c+" ")
		}

		resp := strings.Join(names, ", ")

		client.Conn.Write([]byte(resp + "\n"))
	}
}
