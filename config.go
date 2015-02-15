// Copyright (c) 2015 Bertrand Janin <b@janin.com>
// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/jessevdk/go-flags"
)

// Cmd is a singleton storing all the command-line parameters.
type Cmd struct {
	ConfigFile string `short:"c" description:"Configuration file" default:"/etc/paglop.conf"`
}

// Cfg is a singleton storing all the config file parameters.
type Cfg struct {
	// In Test-mode, this program will not attempt to communicate with any
	// external systems (e.g. SQS and will print everything to stdout).
	// Additionally, all delays are reduced to a minimum to speed up the
	// test suite.
	TestMode bool

	// IRCNickname is the nickname of the bot, passed upon connction.
	IRCNickname string

	// IRCServer is the hostname and port of the IRC server.
	IRCServer string

	// Channels is the list of channels to auto-matically join.
	Channels []string

	// Any chatter from these nicks will be dropped (other bots).
	Ignore []string

	// Where to find the alias file. Will use the local alias file found in
	// the current directory by default.
	AliasFilePath string

	// Where to find the minions file.
	MinionsFilePath string

	// If defined, start a web server to list the aliases (e.g. :8989)
	HTTPServerAddress string
}

var (
	cfg = Cfg{}
	cmd = Cmd{}
)

// GetAutoJoinChannels returns a list of all the auto-join channels (all unique
// configured channels and debug channels).
func (cfg *Cfg) GetAutoJoinChannels() []string {
	channels := make(StringSet, 0)

	for _, name := range cfg.Channels {
		channels.Add(name)
	}

	return channels.Array()
}

// Look in the current directory for an config.json file.
func parseConfigFile() error {
	file, err := os.Open(cmd.ConfigFile)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	if cfg.IRCNickname == "" {
		return errors.New("'IRCNickname' is not defined")
	}

	if cfg.IRCServer == "" {
		return errors.New("'IRCServer' is not defined")
	}

	return nil
}

// Parse the command line arguments and populate the global cmd struct.
func parseCommandLine() {
	flagParser := flags.NewParser(&cmd, flags.PassDoubleDash)
	_, err := flagParser.Parse()
	if err != nil {
		println("command line error: " + err.Error())
		flagParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}
}
