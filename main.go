// Copyright (c) 2015 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"log"
	"os"
	"strings"
)

// These channels represent the lines coming and going to the IRC server (from
// the perspective of this program).
var (
	IRCIncoming   = make(chan string)
	IRCOutgoing   = make(chan string)
	IRCDisconnect = make(chan string)
)

var (
	chain *Chain
)

func handleMsg(msg *Message) {
	output := chain.Generate(15)
	if strings.HasPrefix(output, "ACTION ") {
		output = output[8:]
		privAction(msg.ReplyTo, output)
	} else {
		IRCPrivMsg(msg.ReplyTo, output)
	}
}

func start() error {
	log.Printf("initialize markov chain...")
	chain = initializeMarkovChain()

	log.Printf("connecting...")
	conn, err := connect()
	if err != nil {
		return err
	}

	go connectionReader(conn, IRCIncoming, IRCDisconnect)
	go connectionWriter(conn, IRCOutgoing)

	autojoin()

	for {
		select {
		case data := <-IRCIncoming:
			// Incoming IRC traffic.
			privmsg := NewPrivMsg(data, cfg.IRCNickname)
			if privmsg == nil {
				continue
			}

			if privmsg.Action {
				chain.AddLine("ACTION " + privmsg.Body)
			} else {
				chain.AddLine(privmsg.Body)
			}

			msg := NewMessageFromPrivMsg(privmsg)
			if msg == nil {
				continue
			}

			handleMsg(msg)
		case data := <-IRCDisconnect:
			// Server has disconnected, we're done.
			log.Printf("Disconnected: %s", data)
			return nil
		}
	}

	return nil
}

func main() {
	parseCommandLine()
	err := parseConfigFile()
	if err != nil {
		log.Fatal("config error: ", err.Error())
	}

	log.Printf("starting %s", cfg.IRCNickname)
	err = start()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
