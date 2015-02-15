// Copyright (c) 2015 Bertrand Janin <b@janin.com>
// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// StringSet is used to store a set of strings since Go does not have sets
type StringSet map[string]bool

// Add a new string to the set.
func (ss StringSet) Add(s string) {
	ss[s] = true
}

// Array returns a string array from all the set elements.
func (ss StringSet) Array() []string {
	list := make([]string, len(ss))
	i := 0

	for s := range ss {
		list[i] = s
		i++
	}

	return list
}
