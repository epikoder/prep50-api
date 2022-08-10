package ijwt

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/cache"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/crypto"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

type (
	JWTClaims struct {
		Audience interface{} `json:"aud"`
	}
	JwtToken struct {
		Access    string    `json:"access"`
		Refresh   string    `json:"refresh"`
		ExpiresAt time.Time `json:"access_expires_at"`
		ExpiresRt time.Time `json:"refresh_expires_at"`
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
		claims := JWTClaims{i}
		var buf = []byte{}
		buf, err = AccessSigner.Sign(claims)
		if err != nil {
			return nil, err
		}
		token.Access = string(buf)
		token.ExpiresAt = time.Now().Add(time.Duration(accessExpires) * time.Minute)
		fmt.Println(token.ExpiresAt, time.Duration(token.ExpiresAt.Unix()))
		if err = cache.Set(fmt.Sprintf("%s.access", key), token.Access, cache.Duration(token.ExpiresAt.Unix())); err != nil {
			return
		}
	}
	{
		claims := JWTClaims{i}
		var buf = []byte{}
		buf, err = RefreshSigner.Sign(claims)
		if err != nil {
			return nil, err
		}
		token.Refresh = string(buf)
		token.ExpiresRt = time.Now().Add(time.Duration(refreshExpires) * time.Minute)
		if err = cache.Set(fmt.Sprintf("%s.refresh", key), token.Refresh, cache.Duration(token.ExpiresRt.Unix())); err != nil {
			return
		}
	}
	return
}
