package types

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/patrickmn/go-cache"
	"time"
	"github.com/pkg/errors"
)

type Window struct {
	Id        int64
	Sum       uint64
	Entries   int64
	Slope     float64
	TTL       time.Duration
	HzCeiling uint64
}

func NewWindow() *Window {
	epoch := time.Unix(0, int64(GetConfig().RateLimits.EpochTime))
	minutesSinceEpoch := int64(time.Now().Sub(epoch).Minutes())
	utils.Debug("Minutes since epoch: ", minutesSinceEpoch)

	return &Window{
		Id:        minutesSinceEpoch,
		Sum:       0,
		Entries:   0,
		Slope:     0,
		HzCeiling: 0,
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

func (this *Window) UnmarshalJSON(bytes []byte) error {
	var jsonMap map[string]interface{}
	error := json.Unmarshal(bytes, &jsonMap)
	if error != nil {
		return error
	}
	if jsonMap["id"] != nil {
		val, ok := jsonMap["id"].(float64)
		if !ok {
			return errors.Errorf("value for field 'id' must be a int")
		}
		this.Id = int64(val)
	}
	if jsonMap["sum"] != nil {
		val, ok := jsonMap["sum"].(float64)
		if !ok {
			return errors.Errorf("value for field 'sum' must be a uint")
		}
		this.Sum = uint64(val)
	}
	if jsonMap["entries"] != nil {
		val, ok := jsonMap["entries"].(float64)
		if !ok {
			return errors.Errorf("value for field 'entries' must be a int")
		}
		this.Entries = int64(val)
	}
	if jsonMap["slope"] != nil {
		val, ok := jsonMap["slope"].(float64)
		if !ok {
			return errors.Errorf("value for field 'slope' must be a float")
		}
		this.Slope = val
	}
	if jsonMap["ttl"] != nil {
		val, ok := jsonMap["entries"].(time.Duration)
		if !ok {
			return errors.Errorf("value for field 'ttl' must be a int")
		}
		this.TTL = val
	}
	if jsonMap["hzCeiling"] != nil {
		val, ok := jsonMap["hzCeiling"].(float64)
		if !ok {
			return errors.Errorf("value for field 'entries' must be a int")
		}
		this.HzCeiling = uint64(val)
	}
	return nil
}

func (this Window) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id        int64				`json:"hash"`
		Sum       uint64			`json:"hash"`
		Entries   int64				`json:"hash"`
		Slope     float64			`json:"hash,omitempty"`
		TTL       time.Duration		`json:"hash,omitempty"`
		HzCeiling uint64			`json:"hash,omitempty"`
	}{
		Id: 		this.Id,
		Sum:     	this.Sum,
		Entries:  	this.Entries,
		Slope:    	this.Slope,
		TTL:  		this.TTL,
		HzCeiling:  this.HzCeiling,
	})
}