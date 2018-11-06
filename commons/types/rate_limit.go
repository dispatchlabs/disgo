package types

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/patrickmn/go-cache"
	"math"
	"time"
)

type RateLimit struct {
	Address 	string
	TxRateLimit *TxRateLimit
	Existing	*AccountRateLimits
}

type TxRateLimit struct {
	Amount 		uint64
	TxHash  	string
}

type AccountRateLimits struct {
	TxHashes	[]string
}

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//  RateLimit
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func NewRateLimit(address, txHash string, amount uint64) (*RateLimit, error) {
	return &RateLimit {
		Address:	address,
		TxRateLimit: &TxRateLimit{amount, txHash},
		Existing: &AccountRateLimits{make([]string, 0)},
	}, nil
}

func (this *RateLimit) Set(window Window, txn *badger.Txn, cache *cache.Cache) error {
	existing, err := GetAccountRateLimit(txn, cache, this.Address)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			utils.Error(err)
		}
	}
	if existing != nil {
		existing.TxHashes = append(existing.TxHashes, this.TxRateLimit.TxHash)
		this.Existing = existing
	} else {
		utils.Debug("Adding new key for account: ", this.Address)
		this.Existing.TxHashes = append(this.Existing.TxHashes, this.TxRateLimit.TxHash)
	}

	this.cache(window, cache)
	err = this.persist(txn)
	if err != nil {
		return err
	}
	return nil
}


func (this *RateLimit) cache(window Window, cache *cache.Cache) {
	cache.Set(getAccountRateLimitKey(this.Address), this.Existing, TransactionCacheTTL)
	cache.Set(getTxRateLimitKey(this.TxRateLimit.TxHash), this.TxRateLimit, window.TTL)
}

func (this *RateLimit) persist(txn *badger.Txn) error {

	err := txn.Set([]byte(getAccountRateLimitKey(this.Address)), []byte(this.Existing.string()))
	if err != nil {
		return err
	}
	err = txn.SetWithTTL([]byte(getTxRateLimitKey(this.TxRateLimit.TxHash)), []byte(this.TxRateLimit.string()), TransactionCacheTTL)
	if err != nil {
		return err
	}
	return nil
}

func (this RateLimit) ToPrettyJson() string {
	bytes, err := json.MarshalIndent(this, "", "  ")
	if err != nil {
		utils.Error("unable to marshal RateLimit", err)
		return ""
	}
	return string(bytes)
}

// Formula: Get the current slope from a curve of Hertz over the last 4 hours
// Multiply the slope against Seconds
// Bound the algorithm by lower of 1 second and a max of 24 hours.
func GetCurrentTTL(cache *cache.Cache, window *Window) {
	// guard - if we're below the base hertz threshold keep TTL at zero
	// base Hz is calculated by TxPerMin * AvgHzPerTxn
	if window.Sum <= uint64(GetConfig().RateLimits.TxPerMinute * GetConfig().RateLimits.AvgHzPerTxn) {
		window.TTL = GetConfig().RateLimits.MinTTL
		return
	}

	var previousTTL = time.Duration(0)

	if previousWindow, ok := ToWindowFromCache(cache, window.Id - 1); !ok {
		previousTTL = 0
	} else {
		previousTTL = time.Duration(math.Max(0, float64(previousWindow.TTL)))
	}
	slopeSeconds := time.Duration(window.Slope * float64(time.Second))

	if window.Slope > 0 {
		utils.Info("slope is over zero")
		window.TTL = previousTTL + slopeSeconds
	} else if window.Slope == 0 {
		utils.Info("slope is zero")
		window.TTL = previousTTL
	} else if window.Slope < 0 {
		utils.Info("slope is less than zero")
		// NOTE: subtracting a negative turns positive so you have to add it
		window.TTL = previousTTL + slopeSeconds
	}

	// bounds
	if window.TTL > GetConfig().RateLimits.MaxTTL {
		window.TTL = GetConfig().RateLimits.MaxTTL
	}
	if window.TTL < GetConfig().RateLimits.MinTTL {
		window.TTL = GetConfig().RateLimits.MinTTL
	}

	utils.Info("Current TTL = ", window.TTL.String())
	utils.Info("Slope = ", window.Slope)
}


//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//  Account
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func getAccountRateLimitKey(address string) string {
	return fmt.Sprintf("table-ratelimit-account%s", address)
}

func CheckMinimumAvailable(txn *badger.Txn, cache *cache.Cache, address string, balance uint64) (uint64, error) {

	totalDeduction, err := CalculateLockedAmount(txn, cache, address)
	if err != nil {
		return uint64(0), err
	}
	utils.Debug("Total Hertz Deduction from account = ", totalDeduction)
	available := balance - totalDeduction
	return available, nil
}

func GetAccountRateLimit(txn *badger.Txn, cache *cache.Cache, address string) (*AccountRateLimits, error) {
	key := getAccountRateLimitKey(address)
	value, ok := cache.Get(key)
	if !ok {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return nil, err
		}
		value, err := item.Value()
		if err != nil {
			return nil, err
		}
		accountRateLimits, err := toAccountRateLimitsFromJson(value)
		return accountRateLimits, nil
	}
	arl := value.(*AccountRateLimits)
	return arl, nil
}

func toAccountRateLimitsFromJson(payload []byte) (*AccountRateLimits, error) {
	arl := &AccountRateLimits{}
	err := json.Unmarshal(payload, arl)
	if err != nil {
		return nil, err
	}
	return arl, nil
}

func (this AccountRateLimits) string() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal AccountRateLimits", err)
		return ""
	}
	return string(bytes)
}

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//  Transaction
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func getTxRateLimitKey(hash string) string {
	return fmt.Sprintf("table-ratelimit-transaction-%s", hash)
}

func GetTxRateLimit(cache *cache.Cache, hash string) (*TxRateLimit, error) {
	key := getTxRateLimitKey(hash)
	value, ok := cache.Get(key)
	if !ok {
		//No need to get from DB since we are relying on TTL in cache to give us the correct list
		return nil, nil
	}
	rateLimit := value.(*TxRateLimit)
	return rateLimit, nil
}

func toTxRateLimitFromJson(payload []byte) (*TxRateLimit, error) {
	txRl := &TxRateLimit{}
	err := json.Unmarshal(payload, txRl)
	if err != nil {
		return nil, err
	}
	return txRl, nil
}

func (this TxRateLimit) string() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal RateLimit", err)
		return ""
	}
	return string(bytes)
}

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//  Helpers
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~


func CalculateLockedAmount(txn *badger.Txn, c *cache.Cache, address string) (uint64, error) {
	acctRateLimit, err := GetAccountRateLimit(txn, c, address)
	if err != nil {
		utils.Error(err)
	}
	var totalDeduction uint64
	if acctRateLimit != nil {
		heldTxs := make([]string, 0)
		for _, hash := range acctRateLimit.TxHashes {
			txrl, err := GetTxRateLimit(c, hash)
			if err != nil {
				if err != badger.ErrKeyNotFound {
					utils.Error(err)
					return totalDeduction, err
				}
			}
			if txrl != nil {
				heldTxs = append(heldTxs, txrl.TxHash)
				totalDeduction += txrl.Amount
				utils.Debug(txrl.string())
			}
		}
		acctRateLimit.TxHashes = heldTxs
		c.Set(getAccountRateLimitKey(address), acctRateLimit, TransactionCacheTTL)
		err := txn.Set([]byte(getAccountRateLimitKey(address)), []byte(acctRateLimit.string()))
		if err != nil {
			return totalDeduction, err
		}
	}
	return totalDeduction, nil
}