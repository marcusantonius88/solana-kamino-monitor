package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"kamino-simulador/internal/service"
)

type PositionHandler struct {
	positionService *service.PositionService
	logger          *log.Logger
}

func NewPositionHandler(positionService *service.PositionService, logger *log.Logger) *PositionHandler {
	return &PositionHandler{
		positionService: positionService,
		logger:          logger,
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *PositionHandler) GetPositions(w http.ResponseWriter, r *http.Request) {
	wallet := r.PathValue("wallet")
	if wallet == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "wallet is required"})
		return
	}

	h.logger.Printf("processing /positions request for wallet=%s", wallet)

	debugEnabled := strings.EqualFold(r.URL.Query().Get("debug"), "true")

	result, err := h.positionService.GetWalletPosition(r.Context(), wallet, debugEnabled)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidWallet):
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		default:
			h.logger.Printf("failed to process wallet=%s: %v", wallet, err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		}
		return
	}

	h.logger.Printf(
		"wallet=%s discovery diagnostics: profiles_tried=%d profiles_failed=%d obligations_found=%d obligations_parsed=%d",
		wallet,
		result.Diagnostics.ProfilesTried,
		result.Diagnostics.ProfilesFailed,
		result.Diagnostics.ObligationsFound,
		result.Diagnostics.ObligationsParsed,
	)

	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
