package ijwt

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/crypto"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

var (
	AccessSigner       *jwt.Signer
	RefreshSigner      *jwt.Signer
	Verifier           jwt.Verifier
	JwtGuardMiddleware iris.Handler
)

type (
	JWTClaims struct {
		Audience interface{}
	}
	JwtToken struct {
		Access    string
		Refresh   string
		ExpiresAt time.Time
		ExpiresRt time.Time
	}
)

func init() {
	secret, err := jwt.LoadPrivateKeyECDSA("jwt.key")
	if err != nil {
		logger.HandleError(err)
		if secret, err = crypto.KeyGen(true); err != nil {
			panic(err)
		}
	}
	AccessSigner = jwt.NewSigner(jwt.ES256, secret, 15*time.Minute)
	Verifier := jwt.NewVerifier(jwt.ES256, secret)
	Verifier.WithDefaultBlocklist()
	JwtGuardMiddleware = Verifier.Verify(func() interface{} {
		return new(JWTClaims)
	})
	RefreshSigner = jwt.NewSigner(jwt.ES256, secret, 7*24*time.Hour)
}

func GenerateToken(user *models.User) (token *JwtToken, err error) {
	token = &JwtToken{}
	{
		claims := JWTClaims{
			Audience: map[string]string{
				"email": user.Email,
			},
		}
		t, err := AccessSigner.Sign(claims)
		if err != nil {
			return nil, err
		}
		token.Access = string(t)
		token.ExpiresAt = time.Now().Add(15 * time.Minute)
	}
	{
		claims := JWTClaims{
			Audience: map[string]string{
				"email": user.Email,
			}}
		t, err := RefreshSigner.Sign(claims)
		if err != nil {
			return nil, err
		}
		token.Refresh = string(t)
		token.ExpiresRt = time.Now().Add(7 * 24 * time.Hour)
	}
	return
}
