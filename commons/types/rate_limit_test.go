package types

import (
	"testing"
	"math"
	"fmt"
)


func TestRateLimit(t *testing.T) {
	//hash1 := "44197cc2241ad63b66039e15a85168857272fe1625ed39999972edcdfcbc1bbd"
	//hertz := uint64(25000)
	//address := "test2"
	//rateLimit, err := NewRateLimit(address, hash1, hertz)
	//if err != nil {
	//	t.Error(err)
	//}
	//txn := db.NewTransaction(true)
	//defer txn.Discard()
	//window := helper.AddHertz(txn, services.GetCache(), hertz);
	//rateLimit.Set(*window, txn, c)
}


func TestRateLimitStorage(t *testing.T) {
	hash1 := "44197cc2241ad63b66039e15a85168857272fe1625ed39999972edcdfcbc1bbd"
	//hash2 := "44197cc2241ad63b66039e15a85168857272fe1625ed39999972edcdfcbc1bbe"
	hertz := uint64(25000)
	address := "test"
	rateLimit, err := NewRateLimit(address, hash1, hertz)
	if err != nil {
		t.Error(err)
	}
	addRateLimit(rateLimit)

	fmt.Printf("%s\n", rateLimit.ToPrettyJson())
}

func addRateLimit(rateLimit *RateLimit) {
	//txn := db.NewTransaction(true)
	//defer txn.Discard()
	//rateLimit.Set(txn, c)
	//
	//txRateLimit, err := GetTxRateLimit(c, rateLimit.TxRateLimit.TxHash)
	//if err != nil {
	//	utils.Error(err)
	//}
	//addrsRateLimit, err := GetAccountRateLimit(txn, c, rateLimit.Address)
	//if err != nil {
	//	utils.Error(err)
	//}
	//fmt.Printf("%s\n", txRateLimit.string())
	//fmt.Printf("%s\n", addrsRateLimit.string())
	//
	//var totalDeduction uint64
	//for _, hash := range addrsRateLimit.TxHashes {
	//	utils.Info("Getting Hash: ", hash)
	//	txrl, err := GetTxRateLimit(c, hash)
	//	if err != nil {
	//		utils.Error(err)
	//	}
	//	if txrl != nil {
	//		totalDeduction += txrl.Amount
	//		utils.Info(txrl.string())
	//	}
	//}
	//utils.Info("\nTotal Hertz Deduction from account = ", totalDeduction)
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