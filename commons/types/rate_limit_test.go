package types

import (
	"fmt"
	"github.com/dispatchlabs/disgo/commons/utils"
	"testing"
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
	//window := helper.AddHertz(txn, cache, hertz);
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

func TestGetCurrentTTL(t *testing.T) {
	hzPerMinute := GetConfig().RateLimits.AvgHzPerTxn * GetConfig().RateLimits.TxPerMinute

	window := NewWindow()
	window.Sum = uint64(hzPerMinute * 2)

	// test upper bounds
	window.Slope = 86401.0
	if GetCurrentTTL(c, window); window.TTL != time.Hour*24 {
		t.Errorf("TTL should have been %d hours, it was %d", time.Hour*24, window.TTL/time.Second)
	}

	window.Slope = 82800.0
	if GetCurrentTTL(c, window); window.TTL != time.Hour*23 {
		t.Errorf("TTL should have been %d, it was %d", time.Hour*23, window.TTL/time.Second)
	}

	window.Slope = 999999.0
	if GetCurrentTTL(c, window); window.TTL != time.Hour*24 {
		t.Errorf("TTL should have been %d hours, it was %d", time.Hour*24, window.TTL/time.Second)
	}

	// test negatives
	window.Slope = -24
	if GetCurrentTTL(c, window); window.TTL != time.Second {
		t.Errorf("TTL should have been one second, it was %d", window.TTL/time.Second)
	}

	// regular cases
	window.Slope = 0
	if GetCurrentTTL(c, window); window.TTL != time.Second {
		t.Errorf("TTL should have been one second, it was %d", window.TTL/time.Second)
	}

	window.Slope = 1
	if GetCurrentTTL(c, window); window.TTL != time.Second {
		t.Errorf("TTL should have been one second, it was %d", window.TTL/time.Second)
	}

	window.Slope = 5
	if GetCurrentTTL(c, window); window.TTL != time.Second*5 {
		t.Errorf("TTL should have been five seconds, it was %d", window.TTL/time.Second)
	}

	// test ratcheting
	previousWindow := NewWindow()
	nextWindow := NewWindow()

	window.Id = 2
	previousWindow.Id = 1
	nextWindow.Id = 3

	// ratchet up twice
	previousWindow.TTL = 10 * time.Second
	previousWindow.Cache(c)
	window.Slope = 10
	nextWindow.Slope = 10
	nextWindow.Sum = hzPerMinute * 2

	if GetCurrentTTL(c, window); window.TTL != time.Second*20 {
		t.Errorf("TTL should have been 20 seconds, it was %d", window.TTL/time.Second)
	}
	window.Cache(c)

	if GetCurrentTTL(c, nextWindow); nextWindow.TTL != time.Second*30 {
		t.Errorf("TTL should have been 30 seconds, it was %d", nextWindow.TTL/time.Second)
	}

	// ratchet down twice
	previousWindow.TTL = 100 * time.Second
	previousWindow.Cache(c)
	window.Slope = -10
	nextWindow.Slope = -10
	nextWindow.Sum = hzPerMinute * 2

	if GetCurrentTTL(c, window); window.TTL != time.Second*90 {
		t.Errorf("TTL should have been 90 seconds, it was %d", window.TTL/time.Second)
	}
	window.Cache(c)

	if GetCurrentTTL(c, nextWindow); nextWindow.TTL != time.Second*80 {
		t.Errorf("TTL should have been 80 seconds, it was %d", nextWindow.TTL/time.Second)
	}

	// ratchet up from below the base
	window.Sum = hzPerMinute / 2
	previousWindow.TTL = time.Second
	previousWindow.Cache(c)
	window.Slope = 10
	nextWindow.Slope = 10
	nextWindow.Sum = hzPerMinute * 2

	if GetCurrentTTL(c, window); window.TTL != time.Second {
		t.Errorf("TTL should have been one second, it was %d", window.TTL/time.Second)
	}
	window.Cache(c)

	if GetCurrentTTL(c, nextWindow); nextWindow.TTL != time.Second*11 {
		t.Errorf("TTL should have been 11 seconds, it was %d", nextWindow.TTL/time.Second)
	}

	// ratchet beyond 24 hours
	window.Sum = hzPerMinute * 2
	previousWindow.TTL = time.Second * 43200 // 12 hours
	previousWindow.Cache(c)
	window.Slope = 50000
	nextWindow.Slope = 50000
	nextWindow.Sum = hzPerMinute * 2

	if GetCurrentTTL(c, window); window.TTL != time.Second*86400 {
		t.Errorf("TTL should have been 86400 seconds, it was %d", window.TTL/time.Second)
	}
	window.Cache(c)

	if GetCurrentTTL(c, nextWindow); nextWindow.TTL != time.Second*86400 {
		t.Errorf("TTL should have been 86400 seconds, it was %d", nextWindow.TTL/time.Second)
	}

	// ratchet down from above 24 hours
	window.Sum = hzPerMinute * 2
	previousWindow.TTL = time.Second * 86400 // 12 hours
	previousWindow.Cache(c)
	window.Slope = -21600
	nextWindow.Slope = -21600
	nextWindow.Sum = hzPerMinute * 2

	if GetCurrentTTL(c, window); window.TTL != time.Second*64800 {
		t.Errorf("TTL should have been 6480 seconds, it was %d", window.TTL/time.Second)
	}
	window.Cache(c)

	if GetCurrentTTL(c, nextWindow); nextWindow.TTL != time.Second*43200 {
		t.Errorf("TTL should have been 43200 seconds, it was %d", nextWindow.TTL/time.Second)
	}

	// ratchet to below zero
	window.Sum = hzPerMinute * 2
	previousWindow.TTL = time.Second * 43200 // 12 hours
	previousWindow.Cache(c)
	window.Slope = -43200
	nextWindow.Slope = -43200
	nextWindow.Sum = hzPerMinute * 2

	if GetCurrentTTL(c, window); window.TTL != time.Second {
		t.Errorf("TTL should have been one second, it was %d", window.TTL/time.Second)
	}
	window.Cache(c)

	if GetCurrentTTL(c, nextWindow); nextWindow.TTL != time.Second {
		t.Errorf("TTL should have been one second, it was %d", nextWindow.TTL/time.Second)
	}
}


func TestConfigSettings(t *testing.T) {
	utils.Info("EpochTime: ", GetConfig().RateLimits.EpochTime)
	utils.Info("NumWindows: ", GetConfig().RateLimits.NumWindows)
	utils.Info("TxPerMinute: ", GetConfig().RateLimits.TxPerMinute)
	utils.Info("AvgHzPerTxn: ", GetConfig().RateLimits.AvgHzPerTxn)
}
