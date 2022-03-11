package LibraDB

import (
	"os"
)

type dal struct {
	file *os.File
}


func newDal(path string) (*dal, error) {
	dal := &dal{}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	dal.file = file
	return dal, nil
}
