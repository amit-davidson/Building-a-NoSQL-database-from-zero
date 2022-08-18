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
	if !tx.write {
		tx.db.rwlock.RUnlock()
		return
	}

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
	_, err := tx.db.writeFreelist()
	if err != nil {
		return err
	}

	tx.dirtyNodes = nil
	tx.pagesToDelete = nil
	tx.allocatedPageNums = nil
	tx.db.rwlock.Unlock()
	return nil
}

func (tx *tx) getRootCollection() *Collection {
	rootCollection := newEmptyCollection()
	rootCollection.root = tx.db.root
	rootCollection.tx = tx
	return rootCollection
}

func (tx *tx) GetCollection(name []byte) (*Collection, error) {
	rootCollection := tx.getRootCollection()
	item, err := rootCollection.Find(name)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, nil
	}

	collection := newEmptyCollection()
	collection.deserialize(item)
	collection.tx = tx
	return collection, nil
}

func (tx *tx) CreateCollection(name []byte) (*Collection, error) {
	if !tx.write {
		return nil, writeInsideReadTxErr
	}

	newCollectionPage, err := tx.db.writeNode(NewEmptyNode())
	if err != nil {
		return nil, err
	}

	newCollection := newEmptyCollection()
	newCollection.name = name
	newCollection.root = newCollectionPage.pageNum
	return tx.createCollection(newCollection)
}

func (tx *tx) DeleteCollection(name []byte) error {
	if !tx.write {
		return writeInsideReadTxErr
	}

	rootCollection := tx.getRootCollection()

	return rootCollection.Remove(name)

}

func (tx *tx) createCollection(collection *Collection) (*Collection, error) {
	collection.tx = tx
	collectionBytes := collection.serialize()

	rootCollection := tx.getRootCollection()
	err := rootCollection.Put(collection.name, collectionBytes.value)
	if err != nil {
		return nil, err
	}

	return collection, nil
}