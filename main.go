// Copyright (c) 2015 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"log"
	"os"
)

// These channels represent the lines coming and going to the IRC server (from
// the perspective of this program).
var (
	IRCIncoming   = make(chan string)
	IRCOutgoing   = make(chan string)
	IRCDisconnect = make(chan string)
)

func start() error {
	log.Printf("initialize markov chain...")
	chain := initializeMarkovChain()

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
			msg := NewMessageFromIRCLine(data)
			if msg != nil {
				output := chain.Generate(15)
				IRCPrivMsg(msg.ReplyTo, output)
			}
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
