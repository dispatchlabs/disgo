package state

//func TestMerkleTree(t *testing.T) {
//	var testTX *types.Transaction
//
//	content := make([]tree.MerkleTreeContent, 0)
//	for i := 0; i < 10; i++ {
//		tx := mockTransaction(t)
//		//content = append(content, tx)
//		if i == 3 {
//			testTX = tx
//		}
//	}
//	merkleTree := NewMerkleTree(content)
//	hash := testTX.CalculateHash()
//	fmt.Printf("Key: %v\n", hash)
//	merkleTree.VerifyTree()
//	//exists := merkleTree.Root.Has(merkleTree, hash)
//	//fmt.Printf("Exists = %b for: %s\n", exists, hash)
//}
//
//func mockTransaction(t *testing.T) *types.Transaction {
//
//	key, _ := crypto.NewKey(rand.Reader)
//
//	tx, err := types.NewTransaction(
//		key.GetPrivateKeyString(),
//		0,
//		key.Address,
//		"d70613f93152c84050e7826c4e2b0cc02c1c3b99",
//		1,
//		0,
//		time.Now().UnixNano(),
//	)
//
//	if err != nil {
//		t.Fatalf("Could not create transaction %s", err.Error())
//	}
//	return tx
//}