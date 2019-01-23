package helper

import (
	"encoding/json"
	"github.com/dispatchlabs/disgo/commons/utils"
	"strings"
	"github.com/dispatchlabs/disgo/commons/types"
	"fmt"
	"github.com/pkg/errors"
)

/*
NEED TO ADDD Transaction replay for state of ledger
 */
var dataMap map[string]map[string]string
var countMap map[string]int64

type SyncStats struct {
	TotalAccountCount		int64
	AccountCount			AccountCount
	TotalTransactionCount	int64
	TransactionCount		TransactionCount
	TotalGossipCount		int64
	GossipCount				GossipCount
}

type AccountCount struct {
	GoodAccountCount	int64
	BadAccountCount		int64
	FubarAccountCount	int64
	SubTotal			int64
}

type TransactionCount struct {
	GoodTransactionCount		int64
	OtherBadTransactionCount	int64
	ContractTransferCount		int64
	BalancOfCount				int64
	SubTotal					int64
}

type GossipCount struct {
	GoodGossipCount	int64
	BadGossipCount	int64
	SubTotal		int64
}

// ToPrettyJson
func (this SyncStats) ToPrettyJson() string {
	bytes, err := json.MarshalIndent(this, "", "  ")
	if err != nil {
		utils.Error("unable to marshal SyncStats", err)
		return ""
	}
	return string(bytes)
}

func (this AccountCount) getSum() int64 {
	return this.GoodAccountCount + this.BadAccountCount + this.FubarAccountCount
}

func (this TransactionCount) getSum() int64 {
	return this.GoodTransactionCount + this.OtherBadTransactionCount + this.ContractTransferCount + this.BalancOfCount
}

func (this GossipCount) getSum() int64 {
	return this.GoodGossipCount + this.BadGossipCount
}

func ValidateTxSync(tx *types.Transaction) bool {
	addToCountMap("goodTransactionCount", tx.Key(), tx.String())
	return true
}

//order of checks matters (fall through)
func ValidateSync(keyBytes []byte, valueBytes []byte) bool {
	key := string(keyBytes)
	value := string(valueBytes)
	if strings.HasPrefix(key, "table-account") {
		addToCountMap("TotalAccountCount", key, value)
		account := &types.Account{}
		if err := json.Unmarshal(valueBytes, account); err == nil {
			addToCountMap("goodAccountCount", key, value)
			fmt.Printf("Account: %s\n\n", account.ToPrettyJson() )
			return true
		} else {
			err = handleInvalidAccount(err, key, value)
			utils.Error(err, fmt.Sprintf("Received value: %s is not a valid JSON: %s\n", key, value))
			return false
		}
	} else if strings.HasPrefix(key, "table-transaction") {
		addToCountMap("TotalTransactionCount", key, value)
		tx := &types.Transaction{}
		if err := json.Unmarshal(valueBytes, tx); err == nil {
			addToCountMap("goodTransactionCount", key, value)
			fmt.Printf("Transaction: %s\n\n", tx.ToPrettyJson() )
			return true
		} else {
			err = handleInvalidTransaction(err, key, value)
			utils.Error(err, fmt.Sprintf("Received value: %s is not a valid JSON: %s\n", key, value))
			return false
		}
	} else if strings.HasPrefix(key, "table-gossip") {
		//addToCountMap("TotalGossipCount", key, value)
		if err := json.Unmarshal(valueBytes, &types.Gossip{}); err == nil {
			//addToCountMap("goodGossipCount", key, value)
			return true
		} else {
			//addToCountMap("badGossipCount", key, value)
			//utils.Error(err, fmt.Sprintf("Received value: %s is not a valid JSON: %s\n", key, value))
			return false
		}
	} else if strings.HasPrefix(key, "table-rateLimit") {
		addToCountMap("TotalRateLimitCount", key, value)
		if err := json.Unmarshal(valueBytes, &types.RateLimit{}); err == nil {
			addToCountMap("goodRateLimitCount", key, value)
			return true
		} else {
			addToCountMap("badRateLimitCount", key, value)
			utils.Error(err, fmt.Sprintf("Received value: %s is not a valid JSON: %s\n", key, value))
			return false
		}
	}
	return true
}

func handleInvalidAccount(err error, key, value string) error {
	if strings.Contains(value, "balance") {
		addToCountMap("badAccountCount", key, value)
		return nil
	} else if strings.Contains(value, "Transaction") {
		addToCountMap("fubarAccountCount", key, value)
		return errors.New(err.Error() +  "  Invalid address transfer: ")
	}
	return nil
}

func handleInvalidTransaction(err error, key, value string) error {
	if strings.Contains(value, "balanceOf") {
		addToCountMap("balancOfCount", key, value)
		return nil
	} else if strings.Contains(value, "transfer") {
		addToCountMap("transferCount", key, value)
		return errors.New(err.Error() +  "  Invalid contract transfer: ")
	} else {
		addToCountMap("otherBadTransactionCount", key, value)
		return errors.New(err.Error() +  "  Invalid other transaction: ")
	}
	return nil
}

func addToCountMap(typ, key, value string) {
	if countMap == nil {
		countMap = make(map[string]int64)
	}
	if dataMap == nil {
		dataMap = make(map[string]map[string]string)
	}
	countMap[typ]++
	if dataMap[typ] == nil {
		dataMap[typ] = make(map[string]string)
	}
	dataMap[typ][key] = value
}

func GetCounts() *SyncStats {
	ac := AccountCount{countMap["goodAccountCount"], countMap["badAccountCount"], countMap["fubarAccountCount"], 0}
	ac.SubTotal = ac.getSum()
	tc := TransactionCount{countMap["goodTransactionCount"], countMap["otherBadTransactionCount"], countMap["transferCount"], countMap["balancOfCount"], 0}
	tc.SubTotal = tc.getSum()
	gc := GossipCount{countMap["goodGossipCount"], countMap["badGossipCount"], 0}
	gc.SubTotal = gc.getSum()

	stats := &SyncStats{countMap["TotalAccountCount"],ac,countMap["TotalTransactionCount"],tc,countMap["TotalGossipCount"],gc}

	bytes, err := json.MarshalIndent(dataMap, "", "  ")
	if err != nil {
		utils.Error("unable to marshal dataMap", err)
	}
	utils.WriteFile(".", "badger-data.json", string(bytes))
	fmt.Printf("DataMap: \n%s", string(bytes))
	return stats
}