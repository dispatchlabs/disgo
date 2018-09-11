package types

import (
	"github.com/dgraph-io/badger"
	"github.com/patrickmn/go-cache"
	"encoding/binary"
	"time"
	"fmt"
	"github.com/dispatchlabs/disgo/commons/utils"
	"encoding/json"
	"math"
)

const (
	EXP_GROWTH = 1.645504582
)

type RateLimit struct {
	Address 	string
	TxRateLimit *TxRateLimit
	Existing	*AccountRateLimits
	Db      	*badger.DB
	Page        string
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

func NewRateLimit(address, txHash, page string, amount uint64) (*RateLimit, error) {
	return &RateLimit {
		Address:	address,
		TxRateLimit: &TxRateLimit{amount, txHash},
		Existing: &AccountRateLimits{make([]string, 0)},
		Page: page,
	}, nil
}

func (this *RateLimit) Set(txn *badger.Txn, cache *cache.Cache) error {
	existing, err := GetAccountRateLimit(txn, cache, this.Address)
	if err != nil {
		utils.Error(err)
	}
	if existing != nil {
		existing.TxHashes = append(existing.TxHashes, this.TxRateLimit.TxHash)
		this.Existing = existing
	} else {
		this.Existing.TxHashes = append(this.Existing.TxHashes, this.TxRateLimit.TxHash)
	}

	this.cache(cache)
	err = this.persist(txn)
	if err != nil {
		return err
	}
	return nil
}

func (this *RateLimit) cache(cache *cache.Cache) {
	cache.Set(getAccountRateLimitKey(this.Address), this.Existing, TransactionCacheTTL)
	cache.Set(getTxRateLimitKey(this.TxRateLimit.TxHash), this.TxRateLimit, this.getCurrentTTL())
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

func (this RateLimit) getCurrentTTL() time.Duration {
	value, err := this.Merge()
	if err != nil {
		utils.Error(err)
	}
	nbrSeconds := math.Pow(float64(value), EXP_GROWTH)
	ttl := time.Duration(nbrSeconds) * time.Second
	utils.Info("Current TTL = ", ttl.String())
	return ttl
}

func (this *RateLimit) Merge() (uint64, error) {
	key := []byte(this.Page)
	m := this.Db.GetMergeOperator(key, add, 200*time.Millisecond)
	defer m.Stop()

	m.Add(uint64ToBytes(1))

	res, err := m.Get()
	if err != nil {
		return 0, err
	}
	result := bytesToUint64(res)
	utils.Info("Current Count = ", result)
	return result, nil
}

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//  Account
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func getAccountRateLimitKey(address string) string {
	return fmt.Sprintf("table-ratelimit-account%s", address)
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

func GetTxRateLimit(txn *badger.Txn, cache *cache.Cache, hash string) (*TxRateLimit, error) {
	key := getTxRateLimitKey(hash)
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
		rateLimit, err := toTxRateLimitFromJson(value)
		return rateLimit, nil
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

func uint64ToBytes(i uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], i)
	return buf[:]
}

func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// Merge function to add two uint64 numbers
func add(existing, new []byte) []byte {
	return uint64ToBytes(bytesToUint64(existing) + bytesToUint64(new))
}

func (this *RateLimit) CalculateAndStore(txn *badger.Txn, c *cache.Cache) {
	this.Set(txn, c)

	addrsRateLimit, err := GetAccountRateLimit(txn, c, this.Address)
	if err != nil {
		utils.Error(err)
	}

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