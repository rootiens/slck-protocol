package main

type Channel struct {
	Name    string
	Clients map[*Client]bool
}

func newChannel(name string) *Channel {
	return &Channel{
		Name:    name,
		Clients: make(map[*Client]bool),
	}
}

func (c *Channel) broadcast(s string, m []byte) {
	msg := append([]byte(s), ": "...)
	msg = append(msg, m...)
	msg = append(msg, '\n')

	for cl := range c.Clients {
		cl.Conn.Write(msg)
	}
}

