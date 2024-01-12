package ijwt

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/cache"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/config"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/crypto"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
	_jwt "github.com/kataras/jwt"
)

type (
	JWTClaims struct {
		Aud interface{} `json:"aud"`
		Sub string      `json:"sub"`
		Id  string      `json:"id"`
		Iss string      `json:"iss"`
		Exp int64       `json:"exp"`
		Iat int64       `json:"iat"`
		Aid string      `json:"aid,omitempty"`
	}
	JwtToken struct {
		Access      string    `json:"access"`
		AccessUUID  string    `json:"-"`
		ExpiresAt   time.Time `json:"access_expires_at"`
		RefreshUUID string    `json:"-"`
		Refresh     string    `json:"refresh"`
		ExpiresRt   time.Time `json:"refresh_expires_at"`
	}
	LoginResponse struct {
		*JwtToken
		User interface{} `json:"user"`
	}
)

var (
	AccessSigner       *jwt.Signer
	RefreshSigner      *jwt.Signer
	Verifier           jwt.Verifier
	JwtGuardMiddleware iris.Handler
	accessExpires      int = 15
	refreshExpires     int = 168
	secret             *ecdsa.PrivateKey
)

func init() {
	var err error
	secret, err = jwt.LoadPrivateKeyECDSA("jwt.key")
	if !logger.HandleError(err) {
		if _, err := crypto.KeyGen(true); err != nil {
			panic(err)
		}
	}
	Verifier := jwt.NewVerifier(jwt.ES256, secret)
	Verifier = Verifier.WithDefaultBlocklist()
	JwtGuardMiddleware = Verifier.Verify(func() interface{} {
		return new(JWTClaims)
	})
}

func InitializeSigners() {
	AccessSigner = jwt.NewSigner(jwt.ES256, secret, time.Duration(accessExpires)*time.Minute)
	RefreshSigner = jwt.NewSigner(jwt.ES256, secret, time.Duration(refreshExpires)*time.Minute)
}

func SetAccessLife(d int) {
	accessExpires = d
}

func SetRefreshLife(d int) {
	refreshExpires = d
}

func GenerateToken(i interface{}, key string) (token *JwtToken, err error) {
	token = &JwtToken{}
	{
		key := fmt.Sprintf("%s.access", key)
		claims := JWTClaims{
			Aud: i,
			Id:  key,
			Sub: "access_token",
			Iss: config.Conf.App.Name,
			Iat: time.Now().Unix(),
		}
		token.AccessUUID = claims.Id
		token.ExpiresAt = time.Now().Add(time.Duration(accessExpires) * time.Minute)
		claims.Exp = token.ExpiresAt.Unix()

		var buf = []byte{}
		buf, err = AccessSigner.Sign(claims)
		if err != nil {
			return nil, err
		}
		token.Access = string(buf)
		if err = cache.Set(key, token.Access, cache.Duration(token.ExpiresAt.Unix())); err != nil {
			return
		}
	}
	{
		key := fmt.Sprintf("%s.refresh", key)
		claims := JWTClaims{
			Aud: i,
			Id:  key,
			Sub: "refresh_token",
			Iss: config.Conf.App.Name,
			Iat: time.Now().Unix(),
		}
		claims.Aid = token.AccessUUID
		token.RefreshUUID = claims.Id
		token.ExpiresRt = time.Now().Add(time.Duration(refreshExpires) * time.Minute)
		claims.Exp = token.ExpiresAt.Unix()

		var buf = []byte{}
		buf, err = RefreshSigner.Sign(claims)
		if err != nil {
			return nil, err
		}
		token.Refresh = string(buf)
		if err = cache.Set(key, token.Refresh, cache.Duration(token.ExpiresRt.Unix())); err != nil {
			return
		}
	}
	return
}

func RefreshToken(i interface{}, c JwtToken, key string) (token *JwtToken, err error) {
	token = &JwtToken{}
	{
		key = fmt.Sprintf("%s.access", key)
		claims := JWTClaims{
			Aud: i,
			Id:  key,
			Sub: "access_token",
			Iss: config.Conf.App.Name,
			Iat: time.Now().Unix(),
		}
		token.AccessUUID = claims.Id
		token.ExpiresAt = time.Now().Add(time.Duration(accessExpires) * time.Minute)
		claims.Exp = token.ExpiresAt.Unix()

		var buf = []byte{}
		buf, err = AccessSigner.Sign(claims)
		if err != nil {
			return nil, err
		}
		token.Access = string(buf)
		token.ExpiresRt = c.ExpiresRt
		token.Refresh = c.Refresh
		if err = cache.Set(key, token.Access, cache.Duration(token.ExpiresAt.Unix())); err != nil {
			return
		}
	}
	return
}

func RefreshTokenWithAuth(i interface{}, key string, t JwtToken) (token *JwtToken, err error) {
	token, err = RefreshToken(i, t, key)
	if err != nil {
		return
	}
	return
}

func Decrypt(s string) (claims *JWTClaims, err error) {
	token, err := _jwt.Decode([]byte(strings.TrimPrefix(s, "Bearer ")))
	if err != nil {
		return
	}
	claims = &JWTClaims{}
	if err = json.Unmarshal(token.Payload, claims); err != nil {
		return
	}
	return
}
