// Copyright (c) 2015 Bertrand Janin <b@janin.com>
// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Connection states
const (
	ConnStateInit            = iota
	ConnStateWaitingForHello = iota
	ConnStateLive            = iota
)

// Connection error codes (RFC 1459)
const (
	ErrCodeNoNicknameGiven  = 431
	ErrCodeErroneusNickname = 432
	ErrCodeNicknameInUse    = 433
	ErrCodeNickCollision    = 436
)

var (
	// ConnectionState is the current state of the connection state machine.
	ConnectionState = ConnStateInit

	// reServerMessage is a regexp to parse IRC server messages.
	reServerMessage = regexp.MustCompile(`^:[^ ]+ ([0-9]{2,4}) ([^ ]+) (.*)`)
)

// Send a command to the IRC server.
func sendLine(conn net.Conn, cmd string) {
	cmd = strings.TrimSpace(cmd)
	log.Printf("> %s", cmd)
	fmt.Fprintf(conn, "%s\r\n", cmd)
}

func parseServerMessageCode(line string) int16 {
	tokens := reServerMessage.FindStringSubmatch(line)
	if tokens == nil {
		return 0
	}

	code, err := strconv.ParseInt(tokens[1], 10, 16)
	if err != nil {
		log.Printf("error: invalid server message: bad code (%s) in: %s",
			err.Error(), line)
		return 0
	}

	if tokens[2] != cfg.IRCNickname {
		log.Printf("error: invalid server message: wrong nickname in: %s",
			line)
		return 0
	}

	return int16(code)
}

// Connect to the selected server and join all the specified channels.
func connect() (net.Conn, error) {
	conn, err := net.Dial("tcp", cfg.IRCServer)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func connectionReader(conn net.Conn, incoming chan string, disconnect chan string) {
	bufReader := bufio.NewReader(conn)

	for {
		data, err := bufReader.ReadString('\n')
		if err == io.EOF {
			disconnect <- "server disconnected"
			break
		}
		if err != nil {
			panic(err)
		}

		data = strings.Trim(data, "\r\n")
		log.Printf("< %s", data)

		switch ConnectionState {
		case ConnStateWaitingForHello:
			// This is the NICK/USER phase, add more underscores to
			// the nick, until we find one available. If we get any
			// message other than a NICK error, we assume the
			// server likes us, we move on to channel joining
			// state.
			code := parseServerMessageCode(data)
			log.Printf("Server Message Code: %d", code)

			switch code {
			case ErrCodeNoNicknameGiven, ErrCodeErroneusNickname, ErrCodeNicknameInUse, ErrCodeNickCollision:
				cfg.IRCNickname = cfg.IRCNickname + "_"
				ConnectionState = ConnStateInit
			case 1:
				ConnectionState = ConnStateLive
				incoming <- data
			default:
			}

		case ConnStateLive:
			// Handle PING request from the server. Without these
			// our bot would time out. Don't push that to the
			// channel.
			if strings.Index(data, "PING :") == 0 {
				r := strings.Replace(data, "PING", "PONG", 1)
				fmt.Fprintf(conn, "%s\r\n", r)
				continue
			}

			incoming <- data
		// Standby
		default:
			time.Sleep(20 * time.Millisecond)
		}
	}
}

func connectionWriter(conn net.Conn, outgoing chan string) {
	for {
		switch ConnectionState {
		case ConnStateInit:
			sendLine(conn, fmt.Sprintf("NICK %s", cfg.IRCNickname))
			sendLine(conn, fmt.Sprintf("USER %s localhost "+
				"127.0.0.1 :%s\r\n", cfg.IRCNickname,
				cfg.IRCNickname))
			ConnectionState = ConnStateWaitingForHello
		case ConnStateLive:
			for msg := range outgoing {
				log.Printf("> %s", msg)
				fmt.Fprintf(conn, "%s\r\n", msg)
			}
		// Standby
		default:
			time.Sleep(20 * time.Millisecond)
		}
	}
}
