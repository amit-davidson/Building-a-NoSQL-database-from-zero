package main

func main() {
	dal, _ := newDal("db.db")
	p := dal.allocateEmptyPage()
	p.num = dal.getNextPage()
	p.data = []byte("data")
	_ = dal.writePage(p)
}
