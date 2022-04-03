package main

import "fmt"

func main() {
	dal, _ := newDal("db.db")
	p := dal.allocateEmptyPage()
	p.num = dal.getNextPage()
	copy(p.data[:], "data")
	_ = dal.writePage(p)

	_ = dal.close()

	restartDal, _ := newDal("db.db")
	newP, _ := restartDal.readPage(p.num)
	fmt.Printf("res is: %s", newP.data[0:4])
}
