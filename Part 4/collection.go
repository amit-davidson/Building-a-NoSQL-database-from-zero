package main

import "bytes"

type Collection struct {
	name []byte
	root pgnum

	dal *dal
}

func newCollection(name []byte, root pgnum) *Collection {
	return &Collection{
		name: name,
		root: root,
	}
}

// Put adds a key to the tree. It finds the correct node and the insertion index and adds the item. When performing the
// search, the ancestors are returned as well. This way we can iterate over them to check which nodes were modified and
// rebalance by splitting them accordingly. If the root has too many items, then a new root of a new layer is
// created and the created nodes from the split are added as children.
func (c *Collection) Put(key []byte, value []byte) error {
	i := newItem(key, value)

	// On first insertion the root node does not exist, so it should be created
	var root *Node
	var err error
	if c.root == 0 {
		root, err = c.dal.writeNode(c.dal.newNode([]*Item{i}, []pgnum{}))
		if err != nil {
			return nil
		}
		c.root = root.pageNum
		return nil
	} else {
		root, err = c.dal.getNode(c.root)
		if err != nil {
			return err
		}
	}

	// Find the path to the node where the insertion should happen
	insertionIndex, nodeToInsertIn, ancestorsIndexes, err := root.findKey(i.key, false)
	if err != nil {
		return err
	}

	// If key already exists
	if nodeToInsertIn.items != nil && insertionIndex < len(nodeToInsertIn.items) && bytes.Compare(nodeToInsertIn.items[insertionIndex].key, key) == 0 {
		nodeToInsertIn.items[insertionIndex] = i
	} else {
		// Add item to the leaf node
		nodeToInsertIn.addItem(i, insertionIndex)
	}
	nodeToInsertIn.writeNode(nodeToInsertIn)

	ancestors, err := c.getNodes(ancestorsIndexes)
	if err != nil {
		return err
	}

	// Rebalance the nodes all the way up. Start From one node before the last and go all the way up. Exclude root.
	for i := len(ancestors) - 2; i >= 0; i-- {
		pnode := ancestors[i]
		node := ancestors[i+1]
		nodeIndex := ancestorsIndexes[i+1]
		if node.isOverPopulated() {
			pnode.split(node, nodeIndex)
		}
	}

	// Handle root
	rootNode := ancestors[0]
	if rootNode.isOverPopulated() {
		newRoot := c.dal.newNode([]*Item{}, []pgnum{rootNode.pageNum})
		newRoot.split(rootNode, 0)

		// commit newly created root
		newRoot, err = c.dal.writeNode(newRoot)
		if err != nil {
			return err
		}

		c.root = newRoot.pageNum
	}

	return nil
}

// Find Returns an item according based on the given key by performing a binary search.
func (c *Collection) Find(key []byte) (*Item, error) {
	n, err := c.dal.getNode(c.root)
	if err != nil {
		return nil, err
	}

	index, containingNode, _, err := n.findKey(key, true)
	if err != nil {
		return nil, err
	}
	if index == -1 {
		return nil, nil
	}
	return containingNode.items[index], nil
}

// getNodes returns a list of nodes based on their indexes (the breadcrumbs) from the root
//           p
//       /       \
//     a          b
//  /     \     /   \
// c       d   e     f
// For [0,1,0] -> p,b,e
func (c *Collection) getNodes(indexes []int) ([]*Node, error) {
	root, err := c.dal.getNode(c.root)
	if err != nil {
		return nil, err
	}

	nodes := []*Node{root}
	child := root
	for i := 1; i < len(indexes); i++ {
		child, err = c.dal.getNode(child.childNodes[indexes[i]])
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, child)
	}
	return nodes, nil
}
