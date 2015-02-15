#!/bin/sh
#
# usage: cleanup < irc.log > output.txt
#

sed 's/^..:.. //' \
	| grep -v '^---' \
	| grep -v '^-!-' \
	| grep -iv 'CTCP' \
	| sed 's/^<[ @]*[a-zA-Z0-9_]*> //' \
	| sed 's/^ \* [^ ]*/ACTION/' \
	| sed 's/^..:.. <[ @]*[a-zA-Z0-9_]*> //' \
	| sed 's/^..:.. <[ @]*[a-zA-Z0-9_]*> //'
