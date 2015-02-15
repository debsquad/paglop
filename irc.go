// Copyright (c) 2015 Bertrand Janin <b@janin.com>
// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"strings"
	"time"
)

// privMsg sends a message to a channel.
func privMsg(channel, msg string) {
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

// privAction sends an action message to a channel.
func privAction(channel, msg string) {
	IRCOutgoing <- fmt.Sprintf("PRIVMSG %s :\x01ACTION %s\x01", channel, msg)
}

// join sends a JOIN command.
func join(channel string) {
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
		join(c)
	}
}
