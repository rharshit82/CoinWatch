package main

import (
	"context"

	database "alert-service/database/sqlc"
)

type Alerter interface {
	// creates an alert and pushes to postgres and redis for indexing it in a sorted set
	Create(ctx context.Context, req CreateAlertRequest) (database.Alert, error)

	// Get all your alerts from postgres
	ReadAll(ctx context.Context, req ReadAllAlertsRequest) ([]database.Alert, error)

	// Filter your alerts by status
	ReadFilter(ctx context.Context, req ReadFilerRequest) ([]database.Alert, error)

	// Update an alert in postgres and redis
	Update(ctx context.Context, req UpdateAlertRequest) (database.Alert, error)

	// Delete an alert from postgres and redis
	Delete(ctx context.Context, req DeleteAlertRequest) error
}

type alert struct {
	cache Cacher
	db    database.Querier
}

func NewAlertService(cache Cacher, db database.Querier) Alerter {
	return &alert{
		cache: cache,
		db:    db,
	}
}


// alert is created in postgres, get alert id from postgres, push alert_id with price to redis sorted sets
func (a *alert) Create(ctx context.Context, req CreateAlertRequest) (database.Alert, error) {
	params := database.CreateAlertParams{
		UserID:    req.UserID,
		Crypto:    req.Currency,
		Price:     req.Price,
		Direction: req.Direction,
	}
	res, err := a.db.CreateAlert(ctx, params)
	if err != nil {
		return database.Alert{}, ErrDuplicateAlert
	}

	err = a.cache.AddAlert(ctx, res.ID, res.Crypto, res.Price, res.Direction)
	if err != nil {
		return database.Alert{}, err
	}

	return res, nil
}

func (a *alert) ReadAll(ctx context.Context, req ReadAllAlertsRequest) ([]database.Alert, error) {
	params := database.GetAllAlertsParams{
		UserID: req.UserID,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	res, err := a.db.GetAllAlerts(ctx, params)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (a *alert) ReadFilter(ctx context.Context, req ReadFilerRequest) ([]database.Alert, error) {
	params := database.GetAlertsByStatusParams{
		UserID: req.UserID,
		Status: req.Status,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	res, err := a.db.GetAlertsByStatus(ctx, params)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (a *alert) Update(ctx context.Context, req UpdateAlertRequest) (database.Alert, error) {
	res, err := a.db.GetAlertByID(ctx, req.AlertID)
	if err != nil {
		return database.Alert{}, ErrAlertNotFound
	}

	if res.UserID != req.UserID {
		return database.Alert{}, ErrNotAuthorized
	}

	params := database.UpdateAlertParams{
		ID:        req.AlertID,
		Crypto:    req.Currency,
		Price:     req.Price,
		Direction: req.Direction,
	}
	res, err = a.db.UpdateAlert(ctx, params)
	if err != nil {
		return database.Alert{}, err
	}

	return res, nil
}

func (a *alert) Delete(ctx context.Context, req DeleteAlertRequest) error {
	res, err := a.db.GetAlertByID(ctx, req.AlertID)
	if err != nil {
		return err
	}

	if res.UserID != req.UserID {
		return ErrNotAuthorized
	}

	params := database.UpdateAlertStatusParams{
		ID:     req.AlertID,
		Status: "deleted",
	}

	return a.db.UpdateAlertStatus(ctx, params)
}
