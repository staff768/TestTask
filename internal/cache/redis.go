package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testtask/internal/models"
	"time"

	"testtask/internal/config"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

func Connect(cfg config.RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {

		return nil, fmt.Errorf("failed to connect to Redis: %w", err)

	}

	return &RedisClient{client: rdb, ctx: context.Background()}, nil
}
func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) SetSubscription(sub *models.Subscription) error {
	subJSON, err := json.Marshal(sub)
	if err != nil {
		return err
	}
	err = r.client.Set(r.ctx, strconv.Itoa(sub.ID), subJSON, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}
func (r *RedisClient) GetSubscription(id int) (*models.Subscription, error) {
	subJSON, err := r.client.Get(r.ctx, strconv.Itoa(id)).Result()
	if err != nil {
		return nil, err
	}
	var sub models.Subscription
	err = json.Unmarshal([]byte(subJSON), &sub)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *RedisClient) DeleteSubscription(id int) error {
	return r.client.Del(r.ctx, strconv.Itoa(id)).Err()
}
