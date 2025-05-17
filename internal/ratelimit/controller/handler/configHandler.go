package handler

import (
	"context"
	"encoding/json"
	"errors"
	"ivanjabrony/cloud-test/internal/logger"
	"ivanjabrony/cloud-test/internal/ratelimit/config"
	"ivanjabrony/cloud-test/internal/ratelimit/dto"
	"net/http"
)

type ConfigHandler struct {
	cfg    *config.Config
	logger *logger.MyLogger
	rl     RateLimitService
}

type RateLimitService interface {
	CreateOrUpdateConfig(ctx context.Context, userConfig *dto.UserConfig) error
}

func NewConfigHandler(cfg *config.Config, logger *logger.MyLogger, rl RateLimitService) (*ConfigHandler, error) {
	if cfg == nil || logger == nil || rl == nil {
		return nil, errors.New("nil values in handler constructor")
	}
	return &ConfigHandler{cfg, logger, rl}, nil
}

func (c *ConfigHandler) UpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	var req dto.UserConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Capacity <= 0 || req.RatePerSec <= 0 {
		http.Error(w, "Capacity and rate must be positive", http.StatusBadRequest)
		return
	}

	err := c.rl.CreateOrUpdateConfig(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{
		"status": "configuration updated",
		"ip":     req.Ip,
	})
	if err != nil {
		http.Error(w, "Error while writing response", http.StatusInternalServerError)
		return
	}
}
