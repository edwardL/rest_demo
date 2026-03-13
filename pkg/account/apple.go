package account

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const AppleAuthKeysUrl = "https://appleid.apple.com/auth/keys"

type AppleConfig struct {
	ClientId string `help:"apple报名" default:""`
}

type AppleAuthJwtKeys struct {
	Keys []AppleAuthJwtKeyItem `json:"keys"`
}

type AppleAuthJwtKeyItem struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type AppleTokenPayload struct {
	jwt.RegisteredClaims
	CHash          string `json:"c_hash"`
	Email          string `json:"email"`
	EmailVerified  bool   `json:"email_verified"`
	AuthTime       int    `json:"auth_time"`
	NonceSupported bool   `json:"nonce_supported"`
}

func (tp *AppleTokenPayload) GetIdentType() IdentityType {
	return IdentityTypeApple
}
func (tp *AppleTokenPayload) GetNickname() string {
	//有些可能没有邮箱，给个默认的名字
	if tp.Email == "" {
		return "apple user"
	}
	return ""
}
func (tp *AppleTokenPayload) GetIdentifier() string {
	return tp.Subject
}
func (tp *AppleTokenPayload) GetAccount() string {
	return tp.Email
}
func (tp *AppleTokenPayload) GetAvatar() string {
	return ""
}
func (tp *AppleTokenPayload) GetCredential() string {
	return ""
}

type ThirdAccountApple struct {
	clientId                 string
	publickeys               map[string]*rsa.PublicKey
	lastUpdatePublicKeysTime time.Time
	mux                      *sync.RWMutex
}

func NewThirdAccountApple(conf AppleConfig) *ThirdAccountApple {
	return &ThirdAccountApple{
		clientId:   conf.ClientId,
		publickeys: make(map[string]*rsa.PublicKey),
		mux:        &sync.RWMutex{},
	}
}

// appleAuthToken apple的token校验
// {"iss":"https://appleid.apple.com","sub":"001394.158dbf65ba01451f8345492e0498dd89.0919","aud":["com.chatie.cn"],"exp":1729046101,"iat":1728959701,"c_hash":"3NT86_VIgxQ7Lrd3nYixxQ","email":"437911893@qq.com","email_verified":true,"auth_time":1728959701,"nonce_supported":true}
func (s *ThirdAccountApple) AuthToken(ctx context.Context, appleToken string) (res AuthPayload, err error) {
	var claims AppleTokenPayload
	token, err := jwt.ParseWithClaims(appleToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return s.appleGetKey(ctx, token.Header["kid"].(string))
	}, jwt.WithIssuer("https://appleid.apple.com"), jwt.WithAudience(s.clientId))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("apple auth token check fail")
	}
	//if !claims.EmailVerified {
	//	logs.Error("apple auth error", appleToken)
	//	return nil, fmt.Errorf("google auth token verify fail")
	//}
	return &claims, nil
}

func (s *ThirdAccountApple) appleGetKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	now := time.Now()
	s.mux.RLock()
	lastUpdate := s.lastUpdatePublicKeysTime
	//会缓存一天
	if len(s.publickeys) > 0 && lastUpdate.Add(time.Hour).After(now) {
		if v, ok := s.publickeys[kid]; ok {
			s.mux.RUnlock()
			return v, nil
		}
		s.mux.RUnlock()
		return nil, fmt.Errorf("apple auth key[kid=%s] not found", kid)
	}
	s.mux.RUnlock()

	resp, err := http.Get(AppleAuthKeysUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get apple auth public keys error")
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	appleKeys := AppleAuthJwtKeys{}
	if err := json.Unmarshal(raw, &appleKeys); err != nil {
		return nil, err
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	var curr *rsa.PublicKey
	for _, item := range appleKeys.Keys {
		s.publickeys[item.Kid] = _genApplePublicKey(item)
		if item.Kid == kid {
			curr = s.publickeys[item.Kid]
		}
	}
	s.lastUpdatePublicKeysTime = now
	if curr == nil {
		return nil, fmt.Errorf("apple auth key[kid=%s] not found", kid)
	}
	return curr, nil
}

func _genApplePublicKey(v AppleAuthJwtKeyItem) *rsa.PublicKey {
	var pubKey rsa.PublicKey
	n_bin, _ := base64.RawURLEncoding.DecodeString(v.N)
	n_data := new(big.Int).SetBytes(n_bin)
	e_bin, _ := base64.RawURLEncoding.DecodeString(v.E)
	e_data := new(big.Int).SetBytes(e_bin)
	pubKey.N = n_data
	pubKey.E = int(e_data.Uint64())
	return &pubKey
}
