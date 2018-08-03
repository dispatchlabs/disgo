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
)

// Authenticate
type Authenticate struct {
	Hash      string
	Time      int64
	Signature string
}

// UnmarshalJSON
func (this *Authenticate) UnmarshalJSON(bytes []byte) error {
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
func (this Authenticate) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Hash      string `json:"hash"`
		Time      int64  `json:"time"`
		Signature string `json:"signature"`
	}{
		Hash:      this.Hash,
		Time:      this.Time,
		Signature: this.Signature,
	})
}

func NewAuthenticate() (*Authenticate, error) {
	authenticate := &Authenticate{Time: utils.ToMicroSeconds(time.Now())}

	// Set hash.
	var err error
	authenticate.Hash, err = authenticate.NewHash()
	if err != nil {
		return nil, err
	}

	// Set signature.

	return authenticate, nil
}

// NewHash
func (this Authenticate) NewHash() (string, error) {
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
func (this Authenticate) NewSignature(privateKey string) (string, error) {
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

// Verify
func (this Node) Verify() error {

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
	address := hex.EncodeToString(crypto.ToAddress(publicKeyBytes))
	if address != this.Address {
		return errors.New("node address does not match the computed address from hash and signature")
	}
	if !crypto.VerifySignature(publicKeyBytes, hashBytes, signatureBytes) {
		return errors.New("invalid signature")
	}

	return nil
}
