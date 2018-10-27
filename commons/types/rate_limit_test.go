package types

import (
	"testing"
	"math"
	"fmt"
	"time"
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

func TestGetCurrentTTL(t *testing.T) {
	window := NewWindow()

	// test upper bounds
	window.Slope = 86401.0
	if ttl := GetCurrentTTL(*window); ttl != time.Hour * 24 {
		t.Errorf("TTL should have been %d hours, it was %d", time.Hour * 24, ttl)
	}

	window.Slope = 82800.0
	if ttl := GetCurrentTTL(*window); ttl != time.Hour * 23 {
		t.Errorf("TTL should have been %d, it was %d", time.Hour * 23, ttl)
	}

	window.Slope = 999999.0
	if ttl := GetCurrentTTL(*window); ttl != time.Hour * 24 {
		t.Errorf("TTL should have been %d hours, it was %d", time.Hour * 24, ttl)
	}

	window.Slope = -24
	if ttl :=GetCurrentTTL(*window); ttl != time.Second {
		t.Errorf("TTL should have been one second, it was %d", ttl)
	}

	window.Slope = 0
	if ttl :=GetCurrentTTL(*window); ttl != time.Second {
		t.Errorf("TTL should have been one second, it was %d", ttl)
	}

	window.Slope = 1
	if ttl :=GetCurrentTTL(*window); ttl != time.Second {
		t.Errorf("TTL should have been one second, it was %d", ttl)
	}

	window.Slope = 5
	if ttl :=GetCurrentTTL(*window); ttl != time.Second * 5 {
		t.Errorf("TTL should have been five seconds, it was %d", ttl)
	}
}

func printValue(value float64) {
	fmt.Printf("%f\n", math.Pow(value, EXP_GROWTH))
}