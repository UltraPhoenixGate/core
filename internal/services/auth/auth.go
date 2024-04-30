package auth

import (
	"errors"
	"ultraphx-core/internal/models"
)

func CheckJwtToken(token string) (bool, error) {
	claims, err := ParseJWEToken(token)
	if err != nil {
		return false, err
	}

	// now we not check the token expiration
	// if claims.Expiry.Time().Before(time.Now()) {
	// 	return false, errors.New("token expired")
	// }

	client := models.Client{
		ID: claims.ClientID,
	}

	if err := client.Query().Find(&client); err != nil {
		return false, err.Error
	}

	if client.Status != models.ClientStatusActive {
		if client.Status == models.ClientStatusPending {
			return false, errors.New("client is pending")
		}
		if client.Status == models.ClientStatusExpired {
			return false, errors.New("client is expired")
		}
		if client.Status == models.ClientStatusDisabled {
			return false, errors.New("client is disabled")
		}
		return false, errors.New("client is not active")
	}

	return true, nil
}
