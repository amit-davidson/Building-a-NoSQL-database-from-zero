package main

import "fmt"

func main() {
	db, _ := Open("Demo7", &Options{MinFillPercent: 0.5, MaxFillPercent: 1.0})

	tx := db.WriteTx()
	collectionName := "Demo7Collection"
	createdCollection, _ := tx.CreateCollection([]byte(collectionName))

	newKey := []byte("key0")
	newVal := []byte("value0")
	_ = createdCollection.Put(newKey, newVal)

	_ = tx.Commit()
	_ = db.Close()

	db, _ = Open("Demo7", &Options{MinFillPercent: 0.5, MaxFillPercent: 1.0})
	tx = db.ReadTx()
	createdCollection, _ = tx.GetCollection([]byte(collectionName))

	item, _ := createdCollection.Find(newKey)

	_ = tx.Commit()
	_ = db.Close()

	fmt.Printf("key is: %s, value is: %s\n", item.key, item.value)
}
