// Copyright (c) 2015 Bertrand Janin <b@janin.com>
// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"strings"
	"time"
)

// IRCPrivMsg send a message to a channel.
func IRCPrivMsg(channel, msg string) {
	lines := strings.Split(msg, "\n")
	for i := 0; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		IRCOutgoing <- fmt.Sprintf("PRIVMSG %s :%s", channel, lines[i])

		// Make test mode faster.
		if cfg.TestMode {
			time.Sleep(50 * time.Millisecond)
		} else {
			time.Sleep(500 * time.Millisecond)
		}
	}
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

// Send an action message to a channel.
func privAction(channel, msg string) {
	IRCOutgoing <- fmt.Sprintf("PRIVMSG %s :\x01ACTION %s\x01", channel, msg)
}

// JoinChannel sends a JOIN command.
func JoinChannel(channel string) {
	IRCOutgoing <- "JOIN " + channel
}

// Auto-join all the configured channels.
func autojoin() {
	// Make test mode faster.
	if cfg.TestMode {
		time.Sleep(50 * time.Millisecond)
	} else {
		time.Sleep(500 * time.Millisecond)
	}

	for _, c := range cfg.GetAutoJoinChannels() {
		JoinChannel(c)
	}
}
