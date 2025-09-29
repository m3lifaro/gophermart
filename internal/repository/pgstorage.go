package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (s *PGStorage) GetUserByLogin(ctx context.Context, login string) (*model.UserDao, error) {
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

func (s *PGStorage) CreateUser(ctx context.Context, user *model.UserDao) error {
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
	user.ID = existedID
	if !isNew {
		return ErrUserExists
	}
	return nil
}

func (s *PGStorage) AddOrder(ctx context.Context, userID int32, orderID string) error {
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

func (s *PGStorage) GetOrders(ctx context.Context, userID int32) ([]model.OrderItem, error) {

	rows, err := s.pool.Query(ctx, `
        SELECT 
            order_id, 
            status,
            added_at,
            accrual
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
	var accrual sql.NullFloat64
	var orders []model.OrderItem
	for rows.Next() {
		var order model.OrderItem
		var pgTime time.Time
		err := rows.Scan(
			&order.Number,
			&order.Status,
			&pgTime,
			&accrual,
		)
		order.UploadedAt = pgTime.Format(time.RFC3339)
		if accrual.Valid {
			order.Accrual = accrual.Float64
		} else {
			order.Accrual = 0 // или nil, если ваше поле pointer
		}
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

func (s *PGStorage) UpdateOrder(ctx context.Context, orderID, status string, amount float64, userID int32) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	s.logger.Debug("Updating order", zap.String("order_id", orderID),
		zap.String("status", status),
		zap.Float64("amount", amount),
		zap.Int32("user_id", userID))
	orderUpdateQuery := `UPDATE user_orders SET status = $1, accrual = $2 WHERE order_id = $3`
	_, err = tx.Exec(ctx, orderUpdateQuery, status, amount, orderID)
	if err != nil {
		s.logger.Error("Failed to update user_orders", zap.String("order_id", orderID), zap.Error(err))
		return fmt.Errorf("failed to update order info: %w", err)
	}

	if amount != 0 {
		balanceUpdateQuery := `UPDATE users SET balance = balance + $1 WHERE id = $2`
		_, err = tx.Exec(ctx, balanceUpdateQuery, amount, userID)
		s.logger.Debug("Updated user balance",
			zap.String("order_id", orderID),
			zap.Float64("amount", amount),
			zap.Int32("user_id", userID))
		if err != nil {
			s.logger.Error("Failed to update user balance", zap.String("order_id", orderID), zap.Error(err))
			return fmt.Errorf("failed to update user balance: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (s *PGStorage) WithdrawBonuses(ctx context.Context, userID int32, orderID string, amount float64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	updateBalance := `
        UPDATE users
        SET balance = balance - $1
        WHERE id = $2 AND balance >= $1
    `
	result, err := tx.Exec(ctx, updateBalance, amount, userID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}
	affected := result.RowsAffected()

	if affected == 0 {
		return ErrInsufficientFunds
	}

	insertWithdrawal := `
        INSERT INTO withdrawals(user_id, order_id, amount)
        VALUES ($1, $2, $3)
    `
	_, err = tx.Exec(ctx, insertWithdrawal, userID, orderID, amount)
	if err != nil {
		return fmt.Errorf("failed to insert withdrawal: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (s *PGStorage) GetBalance(ctx context.Context, userID int32) (*model.UserBalance, error) {
	var balance model.UserBalance
	var amount sql.NullFloat64

	err := s.pool.QueryRow(ctx, `
		SELECT u.balance, uam.amount
		FROM (SELECT balance, id FROM users WHERE id = $1) u
		LEFT OUTER JOIN (
			SELECT sum(amount) as amount, user_id FROM withdrawals WHERE user_id = $1 GROUP BY user_id
		) uam ON u.id = uam.user_id;
    `, userID).
		Scan(
			&balance.Current,
			&amount,
		)
	if amount.Valid {
		balance.Withdrawn = amount.Float64
	} else {
		balance.Withdrawn = 0
	}
	if err != nil {
		s.logger.Error("Failed to get user balance",
			zap.Int32("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	return &balance, nil
}

func (s *PGStorage) GetWithdrawals(ctx context.Context, userID int32) ([]model.WithdrawItem, error) {
	rows, err := s.pool.Query(ctx, `
        SELECT 
            order_id, 
            processed_at,
            amount
        FROM withdrawals 
        WHERE user_id = $1
        ORDER BY processed_at DESC
    `, userID)

	if err != nil {
		s.logger.Error("Failed to get user withdrawals",
			zap.Int32("user_id", userID),
			zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	var amount sql.NullFloat64
	var withdrawals []model.WithdrawItem
	for rows.Next() {
		var withdrawal model.WithdrawItem
		var pgTime time.Time
		err := rows.Scan(
			&withdrawal.Order,
			&pgTime,
			&amount,
		)
		withdrawal.ProcessedAt = pgTime.Format(time.RFC3339)
		if amount.Valid {
			withdrawal.Sum = amount.Float64
		} else {
			withdrawal.Sum = 0
		}
		if err != nil {
			s.logger.Error("Failed to scan withdrawal row", zap.Error(err))
			continue
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error during rows iteration", zap.Error(err))
		return nil, err
	}

	return withdrawals, nil
}
