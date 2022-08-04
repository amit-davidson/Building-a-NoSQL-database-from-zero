package main

import "fmt"

func main() {
	dal, _ := newDal("./mainTest")

	node, _ := dal.getNode(dal.root)
	node.dal = dal
	index, containingNode, _ := node.findKey([]byte("Key1"))
	res := containingNode.items[index]

	fmt.Printf("key is: %s, value is: %s\n", res.key, res.value)
	_ = dal.close()
}
