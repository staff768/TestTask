package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"testtask/internal/cache"
	"testtask/internal/config"
	"testtask/internal/models"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type SubscriptionRepository struct {
	db     *sql.DB
	logger *logrus.Logger
	cache  *cache.RedisClient
}

func NewSubscriptionRepository(db *sql.DB, logger *logrus.Logger, cacheClient *cache.RedisClient) *SubscriptionRepository {
	return &SubscriptionRepository{db: db, logger: logger, cache: cacheClient}
}

func Connect(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func (r *SubscriptionRepository) InsertSubscription(sub *models.Subscription) (int, error) {
	var id int
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	var endDate interface{}
	if sub.EndDate != nil {
		endDate = *sub.EndDate
	} else {
		endDate = nil
	}

	if err := r.db.QueryRow(
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		endDate,
	).Scan(&id); err != nil {
		r.logger.WithError(err).Error("Failed to create subscription")
		return 0, fmt.Errorf("failed to create subscription: %w", err)
	}
	r.logger.WithField("subscription_id", id).Info("Subscription created successfully")
	if r.cache != nil {
		sub.ID = id
		if err := r.cache.SetSubscription(sub); err != nil {
			r.logger.WithError(err).Warn("failed to set subscription in cache")
		}
	}
	return id, nil
}
func (r *SubscriptionRepository) GetSubscriptionByID(id int) (*models.Subscription, error) {
	if r.cache != nil {
		if sub, err := r.cache.GetSubscription(id); err == nil && sub != nil {
			r.logger.WithField("subscription_id", id).Info("Subscription loaded from cache")
			return sub, nil
		} else if err != nil {
			r.logger.WithError(err).WithField("subscription_id", id).Debug("Cache lookup failed; falling back to DB")
		}
	}
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions
		WHERE id = $1`

	sub := &models.Subscription{}
	var endDate sql.NullTime
	if err := r.db.QueryRow(query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&endDate,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("subscription not found")
		}
		r.logger.WithError(err).WithField("subscription_id", id).Error("Failed to get subscription")
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	if endDate.Valid {
		t := endDate.Time
		sub.EndDate = &t
	}
	if r.cache != nil {
		if err := r.cache.SetSubscription(sub); err != nil {
			r.logger.WithError(err).Warn("failed to set subscription in cache")
		} else {
			r.logger.WithField("subscription_id", sub.ID).Debug("Subscription cached after DB load")
		}
	}
	r.logger.WithField("subscription_id", id).Info("Subscription loaded from database")
	return sub, nil
}
func (r *SubscriptionRepository) DeleteSubscription(id int) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		r.logger.WithError(err).WithField("subscription_id", id).Error("Failed to delete subscription")
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}
	if r.cache != nil {
		if err := r.cache.DeleteSubscription(id); err != nil {
			r.logger.WithError(err).Warn("failed to delete subscription from cache")
		} else {
			r.logger.WithField("subscription_id", id).Debug("Subscription removed from cache")
		}
	}
	r.logger.WithField("subscription_id", id).Info("Subscription deleted successfully")
	return nil
}
func (r *SubscriptionRepository) UpdateSubscription(subscription *models.Subscription) error {
	if r.cache != nil {
		if err := r.cache.DeleteSubscription(subscription.ID); err != nil {
			r.logger.WithError(err).Warn("failed to delete subscription from cache during update")
		}
		r.logger.WithField("subscription_id", subscription.ID).Info("Subscription removed from cache")
	}
	query := `
		UPDATE subscriptions
		SET service_name = $2, price = $3, start_date = $4, end_date = $5, user_id = $6
		WHERE id = $1
		RETURNING id`

	var endDate interface{}
	if subscription.EndDate != nil {
		endDate = *subscription.EndDate
	} else {
		endDate = nil
	}

	var returnedId int
	if err := r.db.QueryRow(
		query,
		subscription.ID,
		subscription.ServiceName,
		subscription.Price,
		subscription.StartDate,
		endDate,
		subscription.UserID,
	).Scan(&returnedId); err != nil {
		r.logger.WithError(err).WithField("subscription_id", subscription.ID).Error("Failed to update subscription")
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	if r.cache != nil {
		if err := r.cache.SetSubscription(subscription); err != nil {
			r.logger.WithError(err).Warn("failed to update subscription in cache")
		} else {
			r.logger.WithField("subscription_id", subscription.ID).Debug("Subscription updated in cache")
		}
	}

	r.logger.WithField("subscription_id", subscription.ID).Info("Subscription updated successfully")
	return nil
}
func (r *SubscriptionRepository) GetAllSubscription() ([]*models.Subscription, error) {

	rows, err := r.db.Query("SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions")
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()
	subs := []*models.Subscription{}

	for rows.Next() {
		s := &models.Subscription{}
		var endDate sql.NullTime
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &endDate); err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		if endDate.Valid {
			t := endDate.Time
			s.EndDate = &t
		}
		subs = append(subs, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return subs, nil
}

func (r *SubscriptionRepository) SumTotalSubscriptions(startDate, endDate *string, userID *string, serviceName *string) (int64, error) {
	builder := models.NewQueryBuilder().
		WithStartDate(startDate).
		WithEndDate(endDate).
		WithUserId(userID).
		WithServiceName(serviceName)

	query, args := builder.BuildQuery()

	var total sql.NullInt64
	if err := r.db.QueryRow(query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("failed to sum subscriptions: %w", err)
	}
	if !total.Valid {
		return 0, nil
	}
	return total.Int64, nil
}
