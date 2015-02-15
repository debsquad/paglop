#!/bin/sh
#
# usage: cleanup irc.log > output.txt
#

sed 's/^..:.. //' \
	| grep -v '^---' \
	| grep -v '^-!-' \
	| sed 's/^<[ @]*[a-zA-Z0-9_]*> //' \
	| sed 's/^ \* [^ ]*/ACTION/'
