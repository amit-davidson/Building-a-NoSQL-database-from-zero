package LibraDB

import (
	"fmt"
	"os"
)

type pgnum uint64

type page struct {
	num  pgnum
	data []byte
}

type dal struct {
	file     *os.File
	pageSize int

	freelistPage pgnum
	*freelist
}

func newDal(path string, pageSize int) (*dal, error) {
	dal := &dal{}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	dal.file = file
	dal.pageSize = pageSize
	return dal, nil
}

func (d *dal) close() error {
	if d.file != nil {
		err := d.file.Close()
		if err != nil {
			return fmt.Errorf("could not close file: %s", err)
		}
		d.file = nil
	}

	return nil
}

func (d *dal) allocateEmptyPage() *page {
	return &page{
		data: make([]byte, d.pageSize),
	}
}

func (d *dal) readPage(pageNum pgnum) (*page, error) {
	p := d.allocateEmptyPage()

	offset := int(pageNum) * d.pageSize
	_, err := d.file.ReadAt(p.data, int64(offset))
	if err != nil {
		return nil, err
	}
	return p, err
}

func (d *dal) writePage(p *page) error {
	offset := int64(p.num) * int64(d.pageSize)
	_, err := d.file.WriteAt(p.data, offset)
	return err
}

func (d *dal) readFreelist() (*freelist, error) {
	p, err := d.readPage(d.freelistPage)
	if err != nil {
		return nil, err
	}

	freelist := newFreelist()
	freelist.deserialize(p.data)
	return freelist, nil
}

func (d *dal) writeFreelist() (*page, error) {
	p := d.allocateEmptyPage()
	p.num = d.getNextPage()
	p.data = d.freelist.serialize()

	err := d.writePage(p)
	if err != nil {
		return nil, err
	}
	d.freelistPage = p.num
	return p, nil
}