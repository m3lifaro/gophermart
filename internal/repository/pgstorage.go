package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m3lifaro/gophermart/internal/model"
	"go.uber.org/zap"
	"time"
)

type PGStorage struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewPGStorage(pool *pgxpool.Pool, logger *zap.Logger) *PGStorage {
	return &PGStorage{
		pool:   pool,
		logger: logger,
	}
}

func (s *PGStorage) GetUserByLogin(login string) (*model.UserDao, error) {
	ctx := context.TODO()
	var user = &model.UserDao{
		Password: "",
		User:     model.User{},
	}
	err := s.pool.QueryRow(ctx, "SELECT id, login, password FROM users WHERE login = $1", login).
		Scan(
			&user.ID,
			&user.Login,
			&user.Password,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		} else {
			s.logger.Error("Failed to get user by login", zap.String("login", login), zap.Error(err))
			return nil, err
		}
	}
	return user, nil
}

func (s *PGStorage) CreateUser(user *model.UserDao) error {
	ctx := context.TODO()
	var existedID int32
	var isNew bool
	err := s.pool.QueryRow(ctx, `
	   INSERT INTO users(login, password) VALUES($1, $2)
	   ON CONFLICT (login) DO UPDATE SET login = EXCLUDED.login -- фиктивное обновление
	   RETURNING id, (xmax = 0) AS is_new
	`, user.Login, user.Password).Scan(&existedID, &isNew)

	if err != nil {
		s.logger.Error("Failed to set user", zap.Any("user", user), zap.Error(err))
		return err
	}
	if !isNew {
		user.ID = existedID
		return ErrUserExists
	}
	return nil
}

func (s *PGStorage) AddOrder(userID int32, orderID string) error {
	ctx := context.TODO()
	var isNew bool
	var ownerID int32
	err := s.pool.QueryRow(ctx, `
	   INSERT INTO user_orders(user_id, order_id) VALUES($1, $2)
	   ON CONFLICT (order_id) DO UPDATE SET order_id = EXCLUDED.order_id -- фиктивное обновление
	   RETURNING user_id, (xmax = 0) AS is_new
	`, userID, orderID).Scan(&ownerID, &isNew)

	if err != nil {
		s.logger.Error("Failed to set user order",
			zap.Int32("user_id", userID),
			zap.String("orderID", orderID),
			zap.Error(err))
		return err
	}
	if !isNew {
		if userID == ownerID {
			return ErrOrderAlreadyProcessed
		}
		return ErrOrderAlreadyProcessedByOther
	}
	return nil
}

func (s *PGStorage) GetOrders(userID int32) ([]model.OrderItem, error) {
	ctx := context.TODO()

	rows, err := s.pool.Query(ctx, `
        SELECT 
            order_id, 
            status,
            added_at
        FROM user_orders 
        WHERE user_id = $1
        ORDER BY added_at DESC
    `, userID)

	if err != nil {
		s.logger.Error("Failed to get user orders",
			zap.Int32("user_id", userID),
			zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var orders []model.OrderItem
	for rows.Next() {
		var order model.OrderItem
		var pgTime time.Time
		err := rows.Scan(
			&order.Number,
			&order.Status,
			&pgTime,
		)
		order.UploadedAt = pgTime.Format(time.RFC3339)
		if err != nil {
			s.logger.Error("Failed to scan order row", zap.Error(err))
			continue
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error during rows iteration", zap.Error(err))
		return nil, err
	}

	return orders, nil
}
