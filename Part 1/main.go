package main

import "os"

func main() {
	dal, _ := newDal("db.db", os.Getpagesize())
	p := dal.allocateEmptyPage()
	p.num = dal.getNextPage()
	p.data = []byte("data")
	_ = dal.writePage(p)
}
