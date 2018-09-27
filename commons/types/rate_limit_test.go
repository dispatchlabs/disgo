package types

import (
	"testing"
	"math"
	"fmt"
	"github.com/dispatchlabs/disgo/commons/utils"
	"time"
)


func TestRateLimit(t *testing.T) {
	rateLimit := RateLimit{Db: db}
	rateLimit.Merge()
}


func TestRateLimitStorage(t *testing.T) {
	hash1 := "44197cc2241ad63b66039e15a85168857272fe1625ed39999972edcdfcbc1bbd"
	hash2 := "44197cc2241ad63b66039e15a85168857272fe1625ed39999972edcdfcbc1bbe"
	page := "page-1"
	address := "test"
	rateLimit, err := NewRateLimit(address, hash1, page, 5)
	rateLimit.Db = db
	if err != nil {
		t.Error(err)
	}
	addRateLimit(rateLimit)
	time.Sleep(3 *time.Second)
	rateLimit, err = NewRateLimit(address, hash2, page,10)
	rateLimit.Db = db
	if err != nil {
		t.Error(err)
	}
	addRateLimit(rateLimit)
	//fmt.Printf("%s\n", rateLimit.ToPrettyJson())

}

func addRateLimit(rateLimit *RateLimit) {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	rateLimit.Set(txn, c)

	txRateLimit, err := GetTxRateLimit(txn, c, rateLimit.TxRateLimit.TxHash)
	if err != nil {
		utils.Error(err)
	}
		addrsRateLimit, err := GetAccountRateLimit(txn, c, rateLimit.Address)
	if err != nil {
		utils.Error(err)
	}
	fmt.Printf("%s\n", txRateLimit.string())
	fmt.Printf("%s\n", addrsRateLimit.string())

	var totalDeduction uint64
	for _, hash := range addrsRateLimit.TxHashes {
		utils.Info("Getting Hash: ", hash)
		txrl, err := GetTxRateLimit(txn, c, hash)
		if err != nil {
			utils.Error(err)
		}
		if txrl != nil {
			totalDeduction += txrl.Amount
			utils.Info(txrl.string())
		}
	}
	utils.Info("\nTotal Hertz Deduction from account = ", totalDeduction)
}

func TestGrowth(t *testing.T) {
	//MaxTTL := 86400.0  //nbr seconds in a day
	//MinTTL := 1
	UppertTxThreshold := 1000.0

	printValue(1.0)
	printValue(10.0)
	printValue(100.0)
	printValue(250.0)
	printValue(500.0)
	printValue(UppertTxThreshold)

}

func printValue(value float64) {
	fmt.Printf("%f\n", math.Pow(value, EXP_GROWTH))
}