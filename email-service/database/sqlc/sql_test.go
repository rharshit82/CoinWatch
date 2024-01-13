package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQL(t *testing.T) {
	postgres, conn, err := NewPostresDB(context.TODO(), "postgresql://postgres:postgres@localhost:5432/crypto-alert?sslmode=disable")
	assert.NoError(t, err)

	defer conn.Close(context.Background())

	email, err := postgres.GetUserEmailByAlertID(context.Background(), 1)
	assert.NoError(t, err)
	t.Log(email)

	params := UpdateAlertStatusParams{
		ID:     1,
		Status: "completed",
	}

	err = postgres.UpdateAlertStatus(context.Background(), params)
	assert.NoError(t, err)
}
