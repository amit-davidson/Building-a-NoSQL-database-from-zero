package main

import "os"

func main() {
	dal, _ := newDal("db.db", os.Getpagesize())
	p := dal.allocateEmptyPage()
	p.num = dal.getNextPage()
	copy(p.data[:], "data")

	_ = dal.writePage(p)
}
