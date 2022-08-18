package main

type tx struct {
	dirtyNodes    map[pgnum]*Node
	pagesToDelete []pgnum

	// new pages allocated during the transaction. They will be released if rollback is called.
	allocatedPageNums []pgnum

	write bool

	db   *DB
}

func newTx(db *DB, write bool) *tx {
	return &tx{
		map[pgnum]*Node{},
		make([]pgnum, 0),
		make([]pgnum, 0),
		write,
		db,
	}
}

func (tx *tx) newNode(items []*Item, childNodes []pgnum) *Node {
	node := NewEmptyNode()
	node.items = items
	node.childNodes = childNodes
	node.pageNum = tx.db.getNextPage()
	node.tx = tx

	node.tx.allocatedPageNums = append(node.tx.allocatedPageNums, node.pageNum)
	return node
}

func (tx *tx) getNode(pageNum pgnum) (*Node, error) {
	if node, ok := tx.dirtyNodes[pageNum]; ok {
		return node, nil
	}

	node, err := tx.db.getNode(pageNum)
	if err != nil {
		return nil, err
	}
	node.tx = tx
	return node, nil
}

func (tx *tx) writeNode(node *Node) *Node {
	tx.dirtyNodes[node.pageNum] = node
	node.tx = tx
	return node
}

func (tx *tx) deleteNode(node *Node) {
	tx.pagesToDelete = append(tx.pagesToDelete, node.pageNum)
}

func (tx *tx) Rollback() {
	tx.dirtyNodes = nil
	tx.pagesToDelete = nil
	for _, pageNum := range tx.allocatedPageNums {
		tx.db.freelist.releasePage(pageNum)
	}
	tx.allocatedPageNums = nil
	tx.db.rwlock.Unlock()
}

func (tx *tx) Commit() error {
	if !tx.write {
		tx.db.rwlock.RUnlock()
		return nil
	}

	for _, node := range tx.dirtyNodes {
		_, err := tx.db.writeNode(node)
		if err != nil {
			return err
		}
	}

	for _, pageNum := range tx.pagesToDelete {
		tx.db.deleteNode(pageNum)
	}
	freelistPage, err := tx.db.writeFreelist()
	if err != nil {
		return err
	}

	newMeta := newEmptyMeta()
	newMeta.root = tx.db.root
	newMeta.freelistPage = freelistPage.num
	_ , err = tx.db.writeMeta(newMeta)
	if err != nil {
		return err
	}

	tx.dirtyNodes = nil
	tx.pagesToDelete = nil
	tx.allocatedPageNums = nil
	tx.db.rwlock.Unlock()
	return nil
}