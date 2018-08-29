package types

import (
	"encoding/json"
	"github.com/pkg/errors"
	"encoding/hex"
	"github.com/dispatchlabs/disgo/commons/utils"
	"bytes"
	"encoding/binary"
	"github.com/dispatchlabs/disgo/commons/crypto"
	"time"
	"fmt"
	"github.com/patrickmn/go-cache"
)

// Authentication
type Authentication struct {
	Address   string
	Hash      string
	Time      int64
	Signature string
}

// UnmarshalJSON
func (this *Authentication) UnmarshalJSON(bytes []byte) error {
	var jsonMap map[string]interface{}
	var ok bool
	error := json.Unmarshal(bytes, &jsonMap)
	if error != nil {
		return error
	}
	if jsonMap["hash"] != nil {
		this.Hash, ok = jsonMap["hash"].(string)
		if !ok {
			return errors.Errorf("value for field 'hash' must be a string")
		}
	}
	if jsonMap["time"] != nil {
		hertz, ok := jsonMap["time"].(float64)
		if !ok {
			return errors.Errorf("value for field 'time' must be a number")
		}
		this.Time = int64(hertz)
	}
	if jsonMap["signature"] != nil {
		this.Signature, ok = jsonMap["signature"].(string)
		if !ok {
			return errors.Errorf("value for field 'signature' must be a string")
		}
	}
	return nil
}

// MarshalJSON
func (this Authentication) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Address   string `json:"address"`
		Hash      string `json:"hash"`
		Time      int64  `json:"time"`
		Signature string `json:"signature"`
	}{
		Address:   this.Address,
		Hash:      this.Hash,
		Time:      this.Time,
		Signature: this.Signature,
	})
}

// String
func (this Authentication) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal authentication", err)
		return ""
	}
	return string(bytes)
}

// Key
func (this Authentication) Key() string {
	return fmt.Sprintf("table-authentication-%s", this.Hash)
}

// Cache
func (this *Authentication) Cache(cache *cache.Cache) {
	cache.Set(this.Key(), this, AuthenticationCacheTTL)
}

// NewHash
func (this Authentication) NewHash() (string, error) {
	var values = []interface{}{
		this.Time,
	}
	buffer := new(bytes.Buffer)
	for _, value := range values {
		err := binary.Write(buffer, binary.LittleEndian, value)
		if err != nil {
			utils.Error("unable to write node bytes to buffer", err)
			return "", err
		}
	}
	hash := crypto.NewHash(buffer.Bytes())
	return hex.EncodeToString(hash[:]), nil
}

// NewSignature
func (this Authentication) NewSignature(privateKey string) (string, error) {
	hashBytes, err := hex.DecodeString(this.Hash)
	if err != nil {
		utils.Error("unable to decode hash", err)
		return "", err
	}
	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		utils.Error("unable to decode privateKey", err)
		return "", err
	}
	signatureBytes, err := crypto.NewSignature(privateKeyBytes, hashBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(signatureBytes), nil
}

// GetDerivedAddress
func (this Authentication) GetDerivedAddress() (string, error) {
	hashBytes, err := hex.DecodeString(this.Hash)
	if err != nil {
		utils.Error("unable to decode hash", err)
		return "", errors.New("unable to decode hash")
	}
	signatureBytes, err := hex.DecodeString(this.Signature)
	if err != nil {
		utils.Error("unable to decode signature", err)
		return "", errors.New("unable to decode signature")
	}
	publicKeyBytes, err := crypto.ToPublicKey(hashBytes, signatureBytes)
	if err != nil {
		utils.Error("unable to generate public key from hash and signature", err)
		return "", errors.New("unable to generate public key from hash and signature")
	}

	// Compute address.
	return hex.EncodeToString(crypto.ToAddress(publicKeyBytes)), nil
}

// Verify
func (this Authentication) Verify(cache *cache.Cache, address string) error {

	// Is this a duplicate authentication?
	_, err := ToAuthenticationFromCache(cache, address)
	if err != ErrNotFound {
		return errors.New("duplicate authentication")
	}

	// Time out?
	elapsedMilliSeconds := utils.ToMilliSeconds(time.Now()) - this.Time
	if elapsedMilliSeconds > 1500 {
		return errors.New("authentication timed out")
	}

	// Hash ok?
	hash, err := this.NewHash()
	if err != nil {
		return errors.New("unable to compute hash")
	}
	if this.Hash != hash {
		return errors.New("invalid hash")
	}
	hashBytes, err := hex.DecodeString(this.Hash)
	if err != nil {
		utils.Error("unable to decode hash", err)
		return errors.New("unable to decode hash")
	}
	signatureBytes, err := hex.DecodeString(this.Signature)
	if err != nil {
		utils.Error("unable to decode signature", err)
		return errors.New("unable to decode signature")
	}
	publicKeyBytes, err := crypto.ToPublicKey(hashBytes, signatureBytes)
	if err != nil {
		utils.Error("unable to generate public key from hash and signature", err)
		return errors.New("unable to generate public key from hash and signature")
	}

	// Derived address from publicKeyBytes match from?
	computedAddress := hex.EncodeToString(crypto.ToAddress(publicKeyBytes))
	if computedAddress != address {
		return errors.New("node address does not match the computed address from hash and signature")
	}
	if !crypto.VerifySignature(publicKeyBytes, hashBytes, signatureBytes) {
		return errors.New("invalid signature")
	}

	// Cache authentication for duplicate attacks.
	this.Cache(cache)

	return nil
}

// ToAuthenticationFromCache -
func ToAuthenticationFromCache(cache *cache.Cache, hash string) (*Authentication, error) {
	value, ok := cache.Get(fmt.Sprintf("table-authentication-%s", hash))
	if !ok {
		return nil, ErrNotFound
	}
	return value.(*Authentication), nil
}

// NewAuthentication
func NewAuthentication() (*Authentication, error) {
	authenticate := &Authentication{Time: utils.ToMilliSeconds(time.Now())}

	// Set hash.
	var err error
	authenticate.Hash, err = authenticate.NewHash()
	if err != nil {
		return nil, err
	}

	// Set signature.
	authenticate.Signature, err = authenticate.NewSignature(GetAccount().PrivateKey)
	if err != nil {
		return nil, err
	}

	// Verify address equals derived address?
	address, err := authenticate.GetDerivedAddress()
	if err != nil {
		return nil, err
	}
	if address != GetAccount().Address {
		return nil, errors.New("node address does not match derived address")
	}

	return authenticate, nil
}
