package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	gorilla_mux "github.com/gorilla/mux"

	"testtask/internal/models"
	"testtask/internal/repository"
	logger "testtask/pkg"
)

var appRepo *repository.SubscriptionRepository

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// CreateSubscriptionHandler godoc
// @Summary Create subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.CreateSubscriptionRequest true "Create Subscription"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /subscription [post]
func CreateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.WithError(err).Warn("invalid json body")
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.ServiceName == "" || req.Price <= 0 || req.UserID == "" || req.StartDate == "" {
		writeError(w, http.StatusBadRequest, "missing required fields")
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	start, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_date")
		return
	}
	var endPtr *time.Time
	if req.EndDate != "" {
		end, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid end_date")
			return
		}
		endPtr = &end
	}
	sub := &models.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   start,
		EndDate:     endPtr,
	}
	id, err := appRepo.InsertSubscription(sub)
	if err != nil {
		logger.Log.WithError(err).Error("failed to create subscription")
		writeError(w, http.StatusInternalServerError, "failed to create subscription")
		return
	}
	logger.Log.Infof("Subscription created successfully with id: %d", id)
	created, err := appRepo.GetSubscriptionByID(id)
	if err != nil {
		sub.ID = id
		writeJSON(w, http.StatusCreated, sub)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// GetSubscriptionByIdHandler godoc
// @Summary Get subscription by id
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /subscription/{id} [get]
func GetSubscriptionByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := gorilla_mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	sub, err := appRepo.GetSubscriptionByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	logger.Log.Infof("Subscription get with id: %d", sub.ID)
	writeJSON(w, http.StatusOK, sub)
}

// UpdateSubscriptionHandler godoc
// @Summary Update subscription by id
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "Subscription ID"
// @Param subscription body models.UpdateSubscriptionRequest true "Update Subscription"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /subscription/{id} [patch]
func UpdateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	vars := gorilla_mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req models.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	existing, err := appRepo.GetSubscriptionByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if req.ServiceName != "" {
		existing.ServiceName = req.ServiceName
	}
	if req.Price != nil {
		existing.Price = *req.Price
	}
	if req.UserID != "" {
		userID, err := uuid.Parse(req.UserID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		existing.UserID = userID
	}
	if req.StartDate != "" {
		t, err := time.Parse("01-2006", req.StartDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid start_date")
			return
		}
		existing.StartDate = t
	}
	if req.EndDate != "" {
		t, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid end_date")
			return
		}
		existing.EndDate = &t
	}
	if err := appRepo.UpdateSubscription(existing); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update")
		return
	}
	logger.Log.Infof("Subscription update with id: %d", id)

	updated, err := appRepo.GetSubscriptionByID(existing.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load updated object")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// DeleteSubscriptionHandler godoc
// @Summary Delete subscription by id
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /subscription/{id} [delete]
func DeleteSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	vars := gorilla_mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := appRepo.DeleteSubscription(id); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	logger.Log.Infof("Subscription deleted with id: %d", id)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// GetAllSubscriptionHandler godoc
// @Summary List subscriptions
// @Tags subscriptions
// @Produce json
// @Success 200 {array} models.Subscription
// @Failure 500 {object} models.ErrorResponse
// @Router /subscription [get]
func GetAllSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	subs, err := appRepo.GetAllSubscription()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list")
		return
	}
	logger.Log.Info("Get All Subscription")
	writeJSON(w, http.StatusOK, subs)
}

// GetSubscriptionsTotalHandler godoc
// @Summary Sum total price of subscriptions
// @Tags subscriptions
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param user_id query string false "User ID (UUID)"
// @Param service_name query string false "Service name"
// @Success 200 {object} models.TotalResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /subscription/total [get]
func GetSubscriptionsTotalHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	start := q.Get("start_date")
	end := q.Get("end_date")
	user := q.Get("user_id")
	service := q.Get("service_name")

	var startPtr, endPtr, userPtr, servicePtr *string
	if start != "" {
		startPtr = &start
	}
	if end != "" {
		endPtr = &end
	}
	if user != "" {
		userPtr = &user
	}
	if service != "" {
		servicePtr = &service
	}

	total, err := appRepo.SumTotalSubscriptions(startPtr, endPtr, userPtr, servicePtr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to calculate total")
		return
	}
	writeJSON(w, http.StatusOK, models.TotalResponse{Total: total})
}
