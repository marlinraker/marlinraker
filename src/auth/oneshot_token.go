package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"github.com/samber/lo"
	"sync"
	"time"
)

type oneshotToken struct {
	buf    []byte
	expire time.Time
}

var (
	oneshotTokens      = make([]oneshotToken, 0)
	oneshotTokensMutex = &sync.RWMutex{}
	encoding           = base32.StdEncoding.WithPadding(base32.NoPadding)
)

func GenerateOneshotToken() (string, error) {
	buf := make([]byte, 20)
	for {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		oneshotTokensMutex.RLock()
		_, found := lo.Find(oneshotTokens, func(token oneshotToken) bool {
			return bytes.Equal(buf, token.buf)
		})
		oneshotTokensMutex.RUnlock()
		if !found {
			break
		}
	}

	token := oneshotToken{
		buf:    buf,
		expire: time.Now().Add(time.Second * 5),
	}

	oneshotTokensMutex.Lock()
	oneshotTokens = append(oneshotTokens, token)
	oneshotTokensMutex.Unlock()

	go func() {
		time.Sleep(time.Second * 5)
		consumeOneshotToken(buf)
	}()
	return encoding.EncodeToString(buf), nil
}

func ConsumeOneshotToken(base32 string) (bool, error) {
	buf, err := encoding.DecodeString(base32)
	if err != nil {
		return false, err
	}
	return consumeOneshotToken(buf), nil
}

func consumeOneshotToken(buf []byte) bool {
	oneshotTokensMutex.Lock()
	defer oneshotTokensMutex.Unlock()

	now := time.Now()
	for i, token := range oneshotTokens {
		if bytes.Equal(token.buf, buf) {
			oneshotTokens = append(oneshotTokens[:i], oneshotTokens[i+1:]...)
			return now.Before(token.expire)
		}
	}
	return false
}
