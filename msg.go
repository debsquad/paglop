// Copyright (c) 2015 Bertrand Janin <b@janin.com>
// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"strings"
)

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

// NewMessageFromPrivMsg allocates a new Message given an existing PrivMsg.
func NewMessageFromPrivMsg(privmsg *PrivMsg) *Message {
	msg := NewMessage()

	if privmsg.Direct {
		msg.Type = MsgTypeIRCPrivate
	} else {
		// Channel messages have to be prefixes with the bot's nick.
		if !privmsg.Addressed {
			return nil
		}
		msg.Type = MsgTypeIRCChannel
	}

	msg.UserID = privmsg.Nick
	msg.Body = privmsg.Body
	msg.ReplyTo = privmsg.ReplyTo

	tokens := strings.Split(msg.Body, " ")
	if len(tokens) > 0 {
		msg.Command = tokens[0]

		if len(tokens) > 1 {
			msg.Args = append(msg.Args, tokens[1:]...)
		}
	}

	return msg
}

// NewMessageFromIRCLine creates a new Message based on the raw IRC line.
func NewMessageFromIRCLine(line string) *Message {
	privmsg := NewPrivMsg(line, cfg.IRCNickname)
	if privmsg == nil {
		// Not a PRIVMSG.
		return nil
	}

	// Check if we should ignore this message.
	for _, ignore := range cfg.Ignore {
		if ignore == privmsg.Nick {
			return nil
		}
	}

	msg := NewMessageFromPrivMsg(privmsg)

	return msg
}
