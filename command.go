package main

type ID int

const (
	REG ID = iota
	JOIN
	LEAVE
	MSG
	CHNS
	USRS
)

type Command struct {
	Id        ID
	Recipient string
	Sender    string
	Body      []byte
}
