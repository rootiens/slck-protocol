package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
)

var (
	DELIMITER = []byte(`\r\n`)
)

type Client struct {
	Conn       net.Conn
	Outbound   chan<- Command
	Register   chan<- *Client
	Deregister chan<- *Client
	Username   string
}

func newClient(conn net.Conn, o chan<- Command, r chan<- *Client, d chan<- *Client) *Client {
	return &Client{
		Conn:       conn,
		Outbound:   o,
		Register:   r,
		Deregister: d,
	}
}

func (c *Client) read() error {
	for {
		msg, err := bufio.NewReader(c.Conn).ReadBytes('\n')

		if err == io.EOF {
			// On connection close, it will deregister client
			c.Deregister <- c
			return nil
		}

		if err != nil {
			return err
		}

		c.Handle(msg)
	}
}

func (c *Client) Handle(message []byte) {
	cmd := bytes.ToUpper(bytes.TrimSpace(bytes.Split(message, []byte(" "))[0]))
	args := bytes.TrimSpace(bytes.TrimPrefix(message, cmd))

	switch string(cmd) {
	case "REG":
		if err := c.Reg(args); err != nil {
			c.Err(err)
		}
	case "JOIN":
		if err := c.Join(args); err != nil {
			c.Err(err)
		}
	case "LEAVE":
		if err := c.Leave(args); err != nil {
			c.Err(err)
		}
	case "MSG":
		if err := c.Msg(args); err != nil {
			c.Err(err)
		}
	case "CHNS":
		c.Chns()
	case "USRS":
		c.Usrs()
	default:
		c.Err(fmt.Errorf("Unknown command %s", cmd))
	}
}

func (c *Client) Reg(args []byte) error {
	u := bytes.TrimSpace(args)

	if u[0] != '@' {
		return fmt.Errorf("Username must begin with @")
	}

	if len(u) == 0 {
		return fmt.Errorf("Username can't be blank")
	}

	c.Username = string(u)
	c.Register <- c

	return nil
}

func (c *Client) Join(args []byte) error {
	channelID := bytes.TrimSpace(args)
	if channelID[0] != '#' {
		return fmt.Errorf("ERR Channel ID must begin with #")
	}

	c.Outbound <- Command{
		Recipient: string(channelID),
		Sender:    c.Username,
		Id:        JOIN,
	}
	return nil
}

func (c *Client) Leave(args []byte) error {
	channelID := bytes.TrimSpace(args)
	if channelID[0] == '#' {
		return fmt.Errorf("ERR channelID must start with '#'")
	}

	c.Outbound <- Command{
		Recipient: string(channelID),
		Sender:    c.Username,
		Id:        LEAVE,
	}
	return nil
}

func (c *Client) Msg(args []byte) error {
	args = bytes.TrimSpace(args)

	if args[0] != '#' && args[0] != '@' {
		return fmt.Errorf("recipient must be a channel ('#name') or user ('@user')")
	}

	recipient := bytes.Split(args, []byte(" "))[0]
	if len(recipient) == 0 {
		return fmt.Errorf("recipient must have a name")
	}

	args = bytes.TrimSpace(bytes.TrimPrefix(args, recipient))

	l := bytes.Split(args, DELIMITER)[0]
	length, err := strconv.Atoi(string(l))

	if err != nil {
		return fmt.Errorf("body length must be present")
	}

	if length == 0 {
		return fmt.Errorf("body length must be at least 1")
	}

	padding := len(l) + len(DELIMITER) // Size of the body length + the delimiter
	body := args[padding : padding+length]

	c.Outbound <- Command{
		Recipient: string(recipient),
		Sender:    c.Username,
		Body:      body,
		Id:        MSG,
	}

	return nil
}

func (c *Client) Chns() {
	c.Outbound <- Command{
		Sender: c.Username,
		Id:     CHNS,
	}
}

func (c *Client) Usrs() {
	c.Outbound <- Command{
		Sender: c.Username,
		Id:     USRS,
	}
}

func (c *Client) Err(e error) {
	c.Conn.Write([]byte("ERR " + e.Error() + "\n"))
}
