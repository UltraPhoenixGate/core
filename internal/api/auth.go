package api

import (
	"net/http"
	"ultraphx-core/internal/models"
	"ultraphx-core/internal/services/auth"
	"ultraphx-core/pkg/resp"
	"ultraphx-core/pkg/validator"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func HandlePluginRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string   `json:"name" validate:"required"`
		Description string   `json:"description" validate:"required"`
		Permissions []string `json:"permissions"`
	}
	if err := validator.ShouldBind(r, &req); err != nil {
		resp.Error(w, "Invalid request")
		return
	}

	clientPermissions := make([]models.Permission, 0, len(req.Permissions))
	for _, p := range req.Permissions {
		permission, err := models.PrasePermission(p)
		if err != nil {
			resp.Error(w, "Invalid permission")
			return
		}
		clientPermissions = append(clientPermissions, permission)
	}

	client := models.Client{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Type:        models.ClientTypePlugin,
		Permissions: clientPermissions,
	}
	err := client.Query().Create(&client).Error
	if err != nil {
		logrus.WithError(err).Error("Failed to create client")
		resp.Error(w, "Failed to create client")
		return
	}

	token, err := auth.CreateJWEToken(auth.JwtPayload{
		ClientID: client.ID,
		Name:     client.Name,
		Type:     client.Type,
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to create token")
		resp.Error(w, "Failed to create token")
		return
	}
	resp.OK(w, resp.H{
		"token": token,
	})
}

func HandlePluginCheckActive(w http.ResponseWriter, r *http.Request) {
	jwtStr := r.Header.Get("Authorization")
	if jwtStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	ok, err := auth.CheckJwtToken(jwtStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	resp.OK(w, resp.H{
		"success": true,
	})
}
