package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/peymanahmadi/payment-risk-guard/internal/domain"
	"github.com/peymanahmadi/payment-risk-guard/internal/usecase"
)

type Handler struct {
	Process *usecase.ProcessTransaction
	Alerts  domain.AlertRepository
	Txns    domain.TransactionRepository
}

func NewHandler(process *usecase.ProcessTransaction, alerts domain.AlertRepository, txns domain.TransactionRepository) *Handler {
	return &Handler{Process: process, Alerts: alerts, Txns: txns}
}

func (h *Handler) PostTransaction(w http.ResponseWriter, r *http.Request) {
	var req createTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := domain.NewTransaction(req.AccountID, req.Amount, req.Currency, req.Country, req.IPAddress, req.PaymentType, req.OccurredAt)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	outcome, err := h.Process.Execute(r.Context(), tx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to process transaction")
		return
	}

	writeJSON(w, http.StatusCreated, toOutcomeResponse(outcome))
}

func (h *Handler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	accountID := q.Get("account_id")
	severity := domain.Severity(q.Get("severity"))
	limit := parseIntOr(q.Get("limit"), 50)
	offset := parseIntOr(q.Get("offset"), 0)

	alerts, err := h.Alerts.List(r.Context(), accountID, severity, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}

	resp := make([]*alertResponse, 0, len(alerts))
	for i := range alerts {
		resp = append(resp, toAlertResponse(&alerts[i]))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetAlert(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid alert id")
		return
	}

	alert, err := h.Alerts.Get(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "alert not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch alert")
		return
	}

	writeJSON(w, http.StatusOK, toAlertResponse(alert))
}

func (h *Handler) ListAccountTransactions(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	q := r.URL.Query()
	limit := parseIntOr(q.Get("limit"), 50)
	offset := parseIntOr(q.Get("offset"), 0)

	txns, err := h.Txns.ListByAccount(r.Context(), accountID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list transactions")
		return
	}

	resp := make([]transactionResponse, 0, len(txns))
	for i := range txns {
		resp = append(resp, toTransactionResponse(&txns[i]))
	}
	writeJSON(w, http.StatusOK, resp)
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func parseIntOr(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
