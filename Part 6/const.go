package main

import "errors"

const (
	nodeHeaderSize = 3
	pageNumSize = 8
)

var writeInsideReadTxErr = errors.New("can't perform a write operation inside a read transaction")
