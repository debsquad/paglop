// Copyright (c) 2015 Bertrand Janin <b@janin.com>
// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// MsgType describes the type of IRC message.
type MsgType int

// Various MsgType value constants.
const (
	MsgTypeUnknown    MsgType = iota
	MsgTypeIRCChannel MsgType = iota
	MsgTypeIRCPrivate MsgType = iota
	MsgTypeExit       MsgType = iota
	MsgTypeFatal      MsgType = iota
)

// Message is the representation of a bot message/command.
type Message struct {
	Type    MsgType
	UserID  string
	Command string
	Body    string
	// In the case of an IRC message, this is a nickname.
	ReplyTo string
	Args    []string
}

// NewMessage allocates a new blank Message.
func NewMessage() *Message {
	msg := &Message{}
	msg.Type = MsgTypeUnknown
	return msg
}

// NewExitMessage allocates a new Exit message.
func NewExitMessage(body string) *Message {
	msg := NewMessage()
	msg.Type = MsgTypeExit
	msg.Body = body
	return msg
}

// NewFatalMessage allocates a new Fatal error message.
func NewFatalMessage(body string) *Message {
	msg := NewMessage()
	msg.Type = MsgTypeFatal
	msg.Body = body
	return msg
}
