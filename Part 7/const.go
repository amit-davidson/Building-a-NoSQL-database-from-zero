package main

import "errors"

const (
	counterSize = 4
	nodeHeaderSize = 3

	collectionSize = 16
	pageNumSize = 8
)

var writeInsideReadTxErr = errors.New("can't perform a write operation inside a read transaction")