package main

import "os"

func main() {
	// initialize db
	dal, _ := newDal("db.db", os.Getpagesize())

	// create a new page
	p := dal.allocateEmptyPage()
	p.num = dal.getNextPage()
	copy(p.data[:], "data")

	// commit it
	_ = dal.writePage(p)
}
