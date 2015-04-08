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
	output := chain.GenerateOnTopic(32, msg.Body)
	if strings.HasPrefix(output, "ACTION ") {
		output = output[8:]
		privAction(msg.ReplyTo, output)
	} else {
		privMsg(msg.ReplyTo, output)
	}
}

func logLine(channel, line string) {
	filename := cfg.MarkovDataPath + "/autolog-" + channel + ".txt"

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		log.Printf("Error opening %s for logging: %s", filename,
			err.Error())
		return
	}
	defer f.Close()

	_, err = f.WriteString(line + "\n")
	if err != nil {
		log.Printf("Error writing to %s for logging: %s", filename,
			err.Error())
		return
	}
}

func start() error {
	log.Printf("initialize markov chain...")
	chain = initializeMarkovChain(cfg.MarkovDataPath)

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
			var line string

			// Incoming IRC traffic.
			privmsg := NewPrivMsg(data, cfg.IRCNickname)
			if privmsg == nil {
				continue
			}

			if privmsg.Action {
				line = "ACTION " + privmsg.Body
			} else {
				line = privmsg.Body
			}

			chain.AddLine(line)
			logLine(privmsg.ReplyTo, line)

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
