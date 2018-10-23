package types

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"time"
	"encoding/json"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dgraph-io/badger"
)

const (
	DispatchEpoch = 1538352000000000000
	AvgWindowSize = 240
)


type Window struct {
	Id              int64
	Sum				uint64
	Entries         int64
	RollingAverage	uint64
	TTL             string
}


func NewWindow() *Window {
	epoch := time.Unix(0, DispatchEpoch)
	minutesSinceEpoch := int64(time.Now().Sub(epoch).Minutes())
	utils.Debug("Minutes since epoch: ", minutesSinceEpoch)

	return &Window{
		Id: minutesSinceEpoch,
		Sum: 0,
		Entries: 0,
		RollingAverage: 0,
	}
}

func (this *Window) Key() string {
	return GetWindowKey(this.Id)
}

func GetWindowKey(id int64) string {
	return fmt.Sprintf("table-ratelimit-window-%d", id)
}

func (this *Window) AddHertz(cache *cache.Cache, hertz uint64) {
	utils.Debug("AddHertz --> ", hertz)
	if hertz <= 0 {
		return
	}
	this.Sum = this.Sum + hertz
	this.Entries++
	//this.RollingAverage = this.Sum / this.Entries
	this.Cache(cache)
}

func (this *Window) GetHertz(cache *cache.Cache) uint64 {
	value, ok := cache.Get(this.Key())
	if !ok {
		return 0
	}
	return value.(uint64)
}

func (this *Window) Persist(txn *badger.Txn) bool {
	err := txn.Set([]byte(this.Key()), []byte(this.String()))
	if err != nil {
		utils.Error(err)
		return false
	}
	return true
}

func (this *Window) Cache(cache *cache.Cache) {
	cache.Set(this.Key(), this, RateLimitAverageTTL)
}

func ToWindowFromCache(cache *cache.Cache, id int64) (*Window, bool) {
	value, ok := cache.Get(GetWindowKey(id))
	if !ok {
		return nil, false
	}
	window := value.(*Window)
	return window, true
}

// ToWindowFromKey
func ToWindowFromKey(txn *badger.Txn, id int64) (*Window, error) {
	item, err := txn.Get([]byte(GetWindowKey(id)))
	if err != nil {
		return nil, err
	}
	value, err := item.Value()
	if err != nil {
		return nil, err
	}
	window, err := ToWindowFromJson(value)
	if err != nil {
		return nil, err
	}
	return window, err
}

// ToWindowFromJson
func ToWindowFromJson(payload []byte) (*Window, error) {
	window := &Window{}
	err := json.Unmarshal(payload, window)
	if err != nil {
		return nil, err
	}
	return window, nil
}

// String
func (this Window) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal Window", err)
		return ""
	}
	return string(bytes)
}

func (this Window) ToPrettyJson() string {
	bytes, err := json.MarshalIndent(this, "", "  ")
	if err != nil {
		utils.Error("unable to marshal Window", err)
		return ""
	}
	return string(bytes)
}
