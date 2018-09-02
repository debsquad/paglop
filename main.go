// Copyright (c) 2015 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/thoj/go-ircevent"
)

// These channels represent the lines coming and going to the IRC server (from
// the perspective of this program).
var (
	IRCIncoming   = make(chan string)
	IRCOutgoing   = make(chan string)
	IRCDisconnect = make(chan string)

	conn *irc.Connection

	// Detect if we are addressed to.
	reAddressed = regexp.MustCompile(`^(\w+)[:,.]*\s*(.*)`)
)

var (
	chain *Chain
)

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

// MessageHandler is called for every single message, it records sentences and
// makes the bot respond if the sentence is addressed at the bot.
func MessageHandler(nick, target, body string) {
	// We will only respond to a user if they address us, also we won't
	// increment the markov chain with what people tell us since it's often
	// gibberish.
	tokens := reAddressed.FindStringSubmatch(body)
	addressed := tokens != nil && tokens[1] == cfg.IRCNickname
	if !addressed {
		addToMarkov(target, body)
		return
	}
	body = tokens[2]

	// Avoid (up|down)votes from generating a chain.
	if body == "++" || body == "--" {
		return
	}

	output := chain.GenerateOnTopic(10, body)

	// Handle possibly generated ACTIONs.
	if strings.HasPrefix(output, "ACTION ") {
		output = output[7:]
		conn.Action(target, output)
	} else {
		conn.Privmsg(target, output)
	}
}

func addToMarkov(target, body string) {
	chain.AddLine(body)
	logLine(target, body)
}

func main() {
	parseCommandLine()
	err := parseConfigFile()
	if err != nil {
		log.Fatal("config error: ", err.Error())
	}

	log.Printf("initialize markov chain...")
	chain = initializeMarkovChain(cfg.MarkovDataPath)

	conn = irc.IRC(cfg.IRCNickname, cfg.IRCNickname)
	conn.VerboseCallbackHandler = true
	conn.Debug = true
	err = conn.Connect(cfg.IRCServer)
	if err != nil {
		log.Fatal(err)
	}

	conn.AddCallback("001", func(e *irc.Event) {
		for _, c := range cfg.GetAutoJoinChannels() {
			conn.Join(c)
		}
	})

	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		MessageHandler(e.Nick, e.Arguments[0], e.Message())
	})
	conn.AddCallback("CTCP_ACTION", func(e *irc.Event) {
		body := "ACTION " + e.Message()
		target := e.Arguments[0]
		addToMarkov(target, body)
	})

	conn.Loop()

	os.Exit(0)
}
