package dvm

/*
func (self *DVMService) evaluateContract(fromAddress crypto.AddressBytes, contractAddress crypto.AddressBytes, root crypto.HashBytes, sHelper *stateHelper) {
	theEthStateDb := sHelper.WAS.EthStateDB
	contractStateObject := theEthStateDb.GetOrNewStateObject(contractAddress)
	contractHash := crypto.NewHash(contractAddress[:])
	stateHash := theEthStateDb.GetState(contractAddress, contractHash)
	trie := theEthStateDb.StorageTrie(contractAddress)

	fmt.Printf("Contract state object --> \n\n"+
		"From Address:  %v\n"+
		"Address:       %v\n"+
		"Address:       %v\n"+
		"Hash:          %v\n"+
		"Hash:          %v\n"+
		"Nonce:         %v\n"+
		"Code:          %v\n"+
		"Code Hash:     %v\n"+
		"Tree Hash:     %v\n"+
		"Root Hash:     %v\n"+
		"StateHash:     %v\n\n",
		fromAddress,
		crypto.AddressBytesToAddressString(fromAddress),
		contractStateObject.Account().Address,
		contractHash,
		crypto.HashBytesToHashString(contractHash),
		contractStateObject.Account().Nonce,
		contractStateObject.Code(theEthStateDb.Database()),
		contractStateObject.Account().CodeHash,
		trie.Hash(),
		root,
		stateHash,
	)

	// iterateTrie(trie.NodeIterator(root.Bytes()), true)
	// fmt.Println("\n")
	//bytes := state.GetState(address, root)
	//s := state.GetOrNewStateObject(root)

}
*/

/*
func iterateTrie(iterator trie.NodeIterator, isRoot bool) {
	path := hex.EncodeToString(iterator.Path())
	if iterator.Leaf() {
		fmt.Printf("\n"+
			"Leaf Node:   %v\n"+
			"With Path:   %v\n"+
			"With Parent: %v\n"+
			"Leaf Key:    %v\n"+
			"Leaf Blob:   %v\n",
			iterator.Hash(), path, iterator.Parent(), iterator.LeafKey(), iterator.LeafBlob())

	} else {
		if isRoot {
			fmt.Printf("\n"+
				"Root Hash:   %v\n"+
				"With Path:   %v\n",
				iterator.Hash(), path)
		} else {
			fmt.Printf("\n"+
				"Node Hash:   %v\n"+
				"With Path:   %v\n"+
				"With Parent: %v\n",
				iterator.Hash(), path, iterator.Parent())
		}
		iterator.Next(true)
		iterateTrie(iterator, false)
	}
}
*/

/*
func (self *DVMService) resetWAS() {
	self.was = &WriteAheadState{
		db:           self.db,
		EthStateDB:   self.EthStateDB.Copy(),
		txIndex:      0,
		totalUsedGas: big.NewInt(0),
		gp:           new(ethereum.GasPool).AddGas(gasLimit.Uint64()),
	}
	// utils.Info("Reset Write Ahead state")
}
*/

/*
func (self *DVMService) commit(stateHelper *VMStateHelper) (crypto.HashBytes, error) {
	//commit all state changes to the database
	root, err := stateHelper.Commit()
	if err != nil {
		utils.Error(fmt.Sprintf("%s Committing WAS", err))

		return root, err
	}

	// reset the write ahead state for the next block
	// with the latest eth state
	// self.EthStateDB = stateHelper.EthStateDB
	utils.Info(fmt.Sprintf("root %s Committed", root.Hex()))

	// self.resetWAS()

	return root, nil
}
*/

/*
	_, err = dvm.commit(stateHelper) // hash

	// var parsedRes *big.Int
	// err = jsonABI.Unpack(&parsedRes, "test", res)
	//utils.Info(fmt.Sprintf("parsed res: %v", parsedRes))

	// if parsedRes.Cmp(expected) != 0 {
	// 	utils.Error(fmt.Sprintf("Result should be %v, not %v", expected, parsedRes))
	// 	return nil, err
	// }
*/

/*
	if execError != nil {
		utils.Error(execError)
		return nil, execError
	}
	var parsedRes string
	err = jsonABI.Unpack(&parsedRes, "getVar5", execResult)
	utils.Info(fmt.Sprintf("DEBUG-CONTRACT-CALL res: %s", parsedRes))
	if err != nil {
		utils.Error(err)
	}
*/

/*
	_, err = dvm.commit(stateHelper) // hashOfTrieRootNode
	if err != nil {
		utils.Error(err)
	}
*/
