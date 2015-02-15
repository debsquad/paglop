// Copyright (c) 2015 Bertrand Janin <b@janin.com>
// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// Copyright 2011 The Go Authors.  All rights reserved.
//
// Based on a Golang example, itself based on the program presented in the
// "Design and Implementation" chapter of The Practice of Programming
// (Kernighan and Pike, Addison-Wesley 1999).  See also Computer Recreations,
// Scientific American 260, 122 - 125 (1989).
//
// A Markov chain algorithm generates text by creating a statistical model of
// potential textual suffixes for a given prefix. Consider this text:
//
// 	I am not a number! I am a free man!
//
// Our Markov chain algorithm would arrange this text into this set of prefixes
// and suffixes, or "chain": (This table assumes a prefix length of two words.)
//
// 	Prefix       Suffix
//
// 	"" ""        I
// 	"" I         am
// 	I am         a
// 	I am         not
// 	a free       man!
// 	am a         free
// 	am not       a
// 	a number!    I
// 	number! I    am
// 	not a        number!
//
// To generate text using this table we select an initial prefix ("I am", for
// example), choose one of the suffixes associated with that prefix at random
// with probability determined by the input statistics ("a"), and then create a
// new prefix by removing the first word from the prefix and appending the
// suffix (making the new prefix is "am a"). Repeat this process until we can't
// find any suffixes for the current prefix or we exceed the word limit. (The
// word limit is necessary as the chain table may contain cycles.)

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Leader is a Markov chain prefix of one or more words.
type Leader []string

// String returns the Leader as a string (for use as a map key).
func (p Leader) String() string {
	return strings.Join(p, " ")
}

// Shift removes the first word from the Leader and appends the given word.
func (p Leader) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

// Unshift removes the last word from the Leader and put the given word first.
func (p Leader) Unshift(word string) {
	copy(p[1:], p[:len(p)-1])
	p[0] = word
}

// Chain contains a map ("chain") of leaders to a list of follower words.
// A leader is a string of leaderLen words joined with spaces.
// A follower word is a single word. A leader can have multiple follower words.
type Chain struct {
	forward   map[string][]string
	backward  map[string][]string
	words     map[string]uint64
	leaderLen int
}

// NewChain returns a new Chain with leaders of leaderLen words.
func NewChain(leaderLen int) *Chain {
	return &Chain{
		forward:   make(map[string][]string),
		backward:  make(map[string][]string),
		words:     make(map[string]uint64),
		leaderLen: leaderLen,
	}
}

// BadLine decides which lines to skip from the input file.
func BadLine(line string) bool {
	// Comments
	if strings.HasPrefix(line, "#") {
		return true
	}

	// Too small to matter.
	if len(line) < 5 {
		return true
	}

	return false
}

// BadWord decide whether we should keep this word.
func BadWord(word string) bool {
	// We don't care for partial quotes.
	numQuotes := strings.Count(word, `"`)
	if numQuotes != 0 && numQuotes != 2 {
		return true
	}

	// We don't care for partial parenthesis either.
	wordLen := len(word)
	if strings.ContainsAny(word, "()") && (word[0] != '(' || word[wordLen-1] != ')') {
		return true
	}

	return false
}

// AddWord adds a new word to the word registry or increase its score.  This
// registry is used to determine the topic of a sentence.
func (chain *Chain) AddWord(word string) {
	chain.words[word] = chain.words[word] + 1
}

// AddLine adds a new line to the markov chain.
func (chain *Chain) AddLine(line string) {
	var a, b, c string

	if BadLine(line) {
		return
	}

	for _, word := range strings.Fields(line) {
		if BadWord(word) {
			continue
		}
		chain.AddWord(word)
		a, b, c = b, c, word
		fKey := a + " " + b
		bKey := b + " " + c
		chain.forward[fKey] = append(chain.forward[fKey], c)
		chain.backward[bKey] = append(chain.backward[bKey], a)
	}
}

// Build reads text from the provided Reader and
// parses it into leaders and suffixes that are stored in Chain.
func (chain *Chain) Build(r io.Reader) {
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			break
		}

		chain.AddLine(line)
	}
}

func (chain *Chain) GetRandomTupleForWord(word string) string {
	var tuples []string

	for t := range chain.forward {
		if strings.Contains(t, word) {
			tuples = append(tuples, t)
		}
	}

	if len(tuples) == 0 {
		return word
	} else if len(tuples) == 1 {
		return tuples[0]
	}

	return tuples[rand.Intn(len(tuples))]
}

func (chain *Chain) GetLessCommonWord(sentence string) string {
	var lowest uint64 = 9999999
	var chosen string

	for _, word := range strings.Fields(sentence) {
		score := chain.words[word]
		if score > 0 && score < lowest {
			lowest = score
			chosen = word
		}
	}

	return chosen
}

// GenerateCore returns a list of at most n words generated from Chain.
func (chain *Chain) GenerateCore(forward bool, start string, n int) []string {
	p := make(Leader, chain.leaderLen)
	if start != "" {
		p = strings.Fields(start)
	}
	var words []string
	words = append(words, p...)
	for i := 0; i < n; i++ {
		var choices []string
		if forward {
			choices = chain.forward[p.String()]
		} else {
			choices = chain.backward[p.String()]
		}
		if len(choices) == 0 {
			break
		}

		next := choices[rand.Intn(len(choices))]

		if forward {
			p.Shift(next)
		} else {
			// This is the code for "Beginning of sentence.
			if next == " " || next == "" {
				break
			}
			p.Unshift(next)
		}

		words = append(words, next)

		// We have at least 15 words and we have a period, let's stop.
		if len(words) > 10 {
			lastchr := next[len(next)-1]
			if lastchr == '.' || lastchr == '!' || lastchr == '?' {
				break
			}
		}
	}

	return words
}

// Generate returns a string of at most n words generated from Chain.
func (chain *Chain) Generate(n int) string {
	words := chain.GenerateCore(true, "", n)
	return strings.Join(words, " ")
}

// GenerateOnTopic returns a string of at most n words generated from Chain
// using the least common word in the provided sentence.
func (chain *Chain) GenerateOnTopic(n int, sentence string) string {
	word := chain.GetLessCommonWord(sentence)
	log.Printf("Chosen word: %s", word)

	tuple := chain.GetRandomTupleForWord(word)
	log.Printf("Chosen tuple: %s", tuple)

	words := chain.GenerateCore(true, tuple, n)
	return strings.Join(words, " ")
}

// GenerateForward generates words forward in the sentence.
func (chain *Chain) GenerateForward(start string, n int) string {
	words := chain.GenerateCore(true, start, n)
	return strings.Join(words, " ")
}

func getReversedArray(words []string) []string {
	reversed := make([]string, len(words))

	for index, word := range words {
		reversed[len(words)-index-1] = word
	}

	return reversed
}

// GenerateBackward builds sentences backward.
func (chain *Chain) GenerateBackward(end string, n int) string {
	words := chain.GenerateCore(false, end, n)
	fmt.Printf("backward: %s\n", words)
	words = getReversedArray(words)
	return strings.Join(words, " ")
}

func initializeMarkovChain() *Chain {
	rand.Seed(time.Now().UnixNano())

	fileInfos, err := ioutil.ReadDir("data")
	if err != nil {
		println("initializeMarkovChain ReadDir: " + err.Error())
		os.Exit(1)
	}

	chain := NewChain(2)

	for _, fileInfo := range fileInfos {
		filename := fileInfo.Name()
		if !strings.HasSuffix(filename, ".txt") {
			continue
		}
		file, err := os.Open("data/" + filename)
		if err != nil {
			println("initializeMarkovChain Open: " + err.Error())
			os.Exit(1)
		}
		chain.Build(file)
		file.Close()
	}

	return chain
}
