package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)



type Cacher interface {
	AddAlert(ctx context.Context, alertID int64, crypto string, price float64, direction bool) error
	GetTargets(ctx context.Context, crypto currency, direction bool, price string) ([]string, error)
}

type Redis struct {
	client *redis.Client
}

func NewRedis(addr string) (Cacher, error) {
	opt, err := redis.ParseURL(addr)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	return &Redis{
		client: client,
	}, nil
}

func (r *Redis) AddAlert(ctx context.Context, alertID int64, crypto string, price float64, direction bool) error {
	key := formKey(crypto, direction)
	err := r.client.ZAdd(ctx, key, redis.Z{
		Score:  price,
		Member: fmt.Sprint(alertID),
	}).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) GetTargets(ctx context.Context, crypto currency, direction bool, price string) ([]string, error) {
    key := formKey(string(crypto), direction)
    var min, max string

    if direction {
        min = "0"
        max = price
    } else {
        min = price
        max = "inf"
    }

    targets, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
        Min: min,
        Max: max,
    }).Result()
    if err != nil {
        return nil, err
    }

	// delete the targets from ache using zrem
	err = r.client.ZRemRangeByScore(ctx, key, min, max).Err()
	if err != nil {
		return nil, err
	}

    return targets, nil
}


// helper function
func formKey(crypto string, direction bool) string {
	if direction {
		return crypto + ":" + "gt"
	}

	return crypto + ":" + "lt"
}