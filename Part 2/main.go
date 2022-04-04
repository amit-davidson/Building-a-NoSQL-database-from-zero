package main

func main() {
	// initialize db
	dal, _ := newDal("db.db")

	// create a new page
	p := dal.allocateEmptyPage()
	p.num = dal.getNextPage()
	copy(p.data[:], "data")

	// commit it
	_ = dal.writePage(p)
	_, _ = dal.writeFreelist()
	// Close the db
	_ = dal.close()

	// We expect the freelist state was saved, so we write to
	// page number 3 and not overwrite the one at number 2
	dal, _ = newDal("db.db")
	p = dal.allocateEmptyPage()
	p.num = dal.getNextPage()
	copy(p.data[:], "data2")
	_ = dal.writePage(p)
}