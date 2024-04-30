package auth

import (
	"encoding/json"
	"time"
	"ultraphx-core/internal/models"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/sirupsen/logrus"
)

type JwtPayload struct {
	ClientID string            `json:"id"`
	Name     string            `json:"name"`
	Type     models.ClientType `json:"type"`
}

type JwtClaims struct {
	jwt.Claims
	JwtPayload
}

func CreateJWEToken(payload JwtPayload) (string, error) {
	claims := JwtClaims{
		Claims: jwt.Claims{
			Subject:   "Token",
			Issuer:    "Auth Service",
			NotBefore: jwt.NewNumericDate(time.Now()),
			Expiry:    jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
		JwtPayload: payload,
	}

	encrypter, err := jose.NewEncrypter(jose.A128GCM, jose.Recipient{
		Algorithm: jose.RSA_OAEP,
		Key:       GetPublicKey(),
	}, nil)

	if err != nil {
		return "", err
	}

	// 创建JWE令牌
	rawToken, err := jwt.Encrypted(encrypter).Claims(claims).Serialize()
	if err != nil {
		return "", err
	}

	return rawToken, nil
}

func RefreshJWEToken(rawToken string) (string, error) {
	claims, err := ParseJWEToken(rawToken)
	if err != nil {
		return "", err
	}

	claims.NotBefore = jwt.NewNumericDate(time.Now())
	claims.Expiry = jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour))

	encrypter, err := jose.NewEncrypter(jose.A128GCM, jose.Recipient{
		Algorithm: jose.RSA_OAEP,
		Key:       GetPublicKey(),
	}, nil)

	if err != nil {
		return "", err
	}

	// 创建JWE令牌
	rawToken, err = jwt.Encrypted(encrypter).Claims(claims).Serialize()
	if err != nil {
		return "", err
	}

	return rawToken, nil
}

func ParseJWEToken(rawToken string) (*JwtClaims, error) {
	token, err := jose.ParseEncrypted(rawToken, []jose.KeyAlgorithm{jose.RSA_OAEP}, []jose.ContentEncryption{jose.A128GCM})
	if err != nil {
		logrus.WithError(err).Error("Failed to parse token")
		return nil, err
	}

	claims := JwtClaims{}
	tokenData, err := token.Decrypt(GetPrivateKey())
	if err != nil {
		logrus.WithError(err).Error("Failed to decrypt token")
		return nil, err
	}

	if err := json.Unmarshal(tokenData, &claims); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal token")
		return nil, err
	}

	return &claims, nil
}
