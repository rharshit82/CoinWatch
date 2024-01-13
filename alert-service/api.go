package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
	"github.com/rs/zerolog"
)

var logger zerolog.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
	Level(zerolog.TraceLevel).
	With().
	Timestamp().
	Caller().
	Logger()

type API struct {
	listenAddr string
	token      Maker
	auth       Auther
	validator  *validator.Validate
	alert      Alerter
}

func NewAPI(listenAddr string, token Maker, auth Auther, validator *validator.Validate, alert Alerter) *API {
	return &API{
		listenAddr: listenAddr,
		token:      token,
		auth:       auth,
		validator:  validator,
		alert:      alert,
	}
}

func (a *API) Run(ctx context.Context) *http.Server {
	mux := chi.NewRouter()

	// public routes
	mux.Group(func(mux chi.Router) {
		mux.Get("/", a.handle(a.root))
		mux.Post("/signup", a.handle(a.signUp))
		mux.Get("/login", a.handle(a.login))
	})

	// private routes
	mux.Route("/alerts", func(mux chi.Router) {
		mux.Post("/create", a.handle(a.authMiddleware(a.createAlert)))
		mux.Get("/read", a.handle(a.authMiddleware(a.readAlert)))
		mux.Get("/read/filter", a.handle(a.authMiddleware(a.readFilterAlert)))
		mux.Put("/update", a.handle(a.authMiddleware(a.updateAlert)))
		mux.Delete("/delete", a.handle(a.authMiddleware(a.deleteAlert)))
	})

	server := &http.Server{
		Addr:    a.listenAddr,
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	return server
}

// Root handler
func (a *API) root(w http.ResponseWriter, r *http.Request) error {
	resp := map[string]string{"message": "ok"}
	return writeJSON(r.Context(), w, http.StatusOK, resp)
}

// Sign Up handler
func (a *API) signUp(w http.ResponseWriter, r *http.Request) error {
	var req SignUpUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return ErrBadRequest
	}

	err = a.validator.Struct(req)
	if err != nil {
		return NewErrValidation(err)
	}

	resp, err := a.auth.SignUp(r.Context(), req)
	if err != nil {
		return err
	}

	return writeJSON(r.Context(), w, http.StatusOK, resp)
}

// Login handler
func (a *API) login(w http.ResponseWriter, r *http.Request) error {
	var req LoginUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return ErrBadRequest
	}

	err = a.validator.Struct(req)
	if err != nil {
		return NewErrValidation(err)
	}

	resp, err := a.auth.Login(r.Context(), req)
	if err != nil {
		return err
	}

	return writeJSON(r.Context(), w, http.StatusOK, resp)
}

// Create Alert handler
func (a *API) createAlert(w http.ResponseWriter, r *http.Request) error {
	var req CreateAlertRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return ErrBadRequest
	}

	err = a.validator.Struct(req)
	if err != nil {
		return NewErrValidation(err)
	}

	resp, err := a.alert.Create(r.Context(), req)
	if err != nil {
		return err
	}

	return writeJSON(r.Context(), w, http.StatusOK, resp)
}

// Read Alert handler
func (a *API) readAlert(w http.ResponseWriter, r *http.Request) error {
	var req ReadAllAlertsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return ErrBadRequest
	}

	err = a.validator.Struct(req)
	if err != nil {
		return NewErrValidation(err)
	}

	resp, err := a.alert.ReadAll(r.Context(), req)
	if err != nil {
		return err
	}

	return writeJSON(r.Context(), w, http.StatusOK, resp)
}

// Filter alerts
func (a *API) readFilterAlert(w http.ResponseWriter, r *http.Request) error {
	var req ReadFilerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return ErrBadRequest
	}

	err = a.validator.Struct(req)
	if err != nil {
		return NewErrValidation(err)
	}

	resp, err := a.alert.ReadFilter(r.Context(), req)
	if err != nil {
		return err
	}

	return writeJSON(r.Context(), w, http.StatusOK, resp)
}

// update alert handler
func (a *API) updateAlert(w http.ResponseWriter, r *http.Request) error {
	var req UpdateAlertRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return ErrBadRequest
	}

	err = a.validator.Struct(req)
	if err != nil {
		return NewErrValidation(err)
	}

	resp, err := a.alert.Update(r.Context(), req)
	if err != nil {
		return err
	}

	return writeJSON(r.Context(), w, http.StatusOK, resp)
}

// Delete Alert handler
func (a *API) deleteAlert(w http.ResponseWriter, r *http.Request) error {
	var req DeleteAlertRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return ErrBadRequest
	}

	err = a.validator.Struct(req)
	if err != nil {
		return NewErrValidation(err)
	}

	err = a.alert.Delete(r.Context(), req)
	if err != nil {
		return err
	}

	return writeJSON(r.Context(), w, http.StatusOK, nil)
}

// centralize error handling
type Handler func(w http.ResponseWriter, r *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func (a *API) handle(next Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), Route, r.URL.Path))
		r = r.WithContext(context.WithValue(r.Context(), Method, r.Method))

		if err := next(w, r); err != nil {
			switch err {
			case ErrBadRequest, ErrNoAuthHeader, ErrInvalidAuthHeader, ErrUnsupportedAuthType, ErrUserAlreadyExists, ErrDuplicateAlert, ErrAlertNotFound:
				writeJSON(r.Context(), w, http.StatusBadRequest, ApiError{Error: err.Error()})

			case ErrNotAuthorized, ErrTokenExpired, ErrInvalidToken:
				writeJSON(r.Context(), w, http.StatusUnauthorized, ApiError{Error: err.Error()})

			default:
				if vErr, ok := err.(*ErrValidation); ok {
					writeJSON(r.Context(), w, http.StatusBadRequest, ApiError{Error: vErr.Error()})
				} else {
					log.Println("critical internal server error:", err)
					writeJSON(r.Context(), w, http.StatusInternalServerError, ApiError{Error: "internal server error"})
				}
			}
		}
	}
}

// middlewares
type Request struct {
	UserID int64 `json:"user_id"`
}

func (a *API) authMiddleware(next Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		authorizationHeader := r.Header.Get("authorization")

		if len(authorizationHeader) == 0 {
			return ErrNoAuthHeader
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			return ErrInvalidAuthHeader
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != "bearer" {
			return ErrUnsupportedAuthType
		}

		accessToken := fields[1]
		payload, err := a.token.Verify(accessToken)
		if err != nil {
			return err
		}

		var req Request
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		err = json.Unmarshal(body, &req)
		if err != nil {
			return ErrBadRequest
		}
		if payload.UserID != req.UserID {
			return ErrNotAuthorized
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))
		return next(w, r)
	}
}

// helper function
func writeJSON(ctx context.Context, w http.ResponseWriter, s int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s)

	// centralized logging
	if apiErr, ok := v.(ApiError); ok {
		logger.Error().
			Int("status", s).
			Str("route", ctx.Value(Route).(string)).
			Str("method", ctx.Value(Method).(string)).
			Str("err", apiErr.Error).
			Send()
	} else {
		logger.Info().
			Int("status", s).
			Str("route", ctx.Value(Route).(string)).
			Str("method", ctx.Value(Method).(string)).
			Send()
	}

	return json.NewEncoder(w).Encode(v)
}
