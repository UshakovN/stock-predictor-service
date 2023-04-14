package storage

import (
	"context"
	"fmt"
	"main/internal/storage/postgres"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/UshakovN/stock-predictor-service/utils"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

const (
	txRetryCount   = 5
	txWaitInterval = 1 * time.Second
)

type queryBuilder interface {
	ToSql() (string, []any, error)
}

type Storage interface {
	PutUser(user *ServiceUser) error
	GetUser(userId, userEmail string) (*ServiceUser, bool, error)
	PutToken(token *RefreshToken) error
	GetToken(tokenId string) (*RefreshToken, bool, error)
	UpdateToken(token *RefreshToken) error
}

type storage struct {
	ctx    context.Context
	client postgres.Client
}

func NewStorage(ctx context.Context, config *postgres.Config) (Storage, error) {
	client, err := postgres.NewClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("cannot create new postgres client: %v", err)
	}
	return &storage{
		ctx:    ctx,
		client: client,
	}, nil
}

func (s *storage) PutUser(user *ServiceUser) error {
	return utils.DoWithRetry(func() error {
		tx, err := s.client.BeginTx(s.ctx, pgx.TxOptions{
			IsoLevel: pgx.Serializable,
		})
		if err != nil {
			return nil
		}
		selectBuilder := sq.
			Select(`COUNT(*)`).
			From(`stock_service_user`).
			Where(sq.Eq{
				`email`: user.Email,
			})
		var usersCount int

		if _, err := s.doQuery(tx, selectBuilder, &usersCount); err != nil {
			return fmt.Errorf("cannot do transaction query: %v", err)
		}
		if usersCount != 0 {
			if err = tx.Rollback(s.ctx); err != nil {
				return fmt.Errorf("cannot rollback transaction: %v", err)
			}
		}
		insertBuilder := sq.
			Insert(`stock_service_user`).
			Columns(
				`user_id`,
				`email`,
				`password_hash`,
				`full_name`,
				`active`,
				`created_at`,
			).
			Values(
				user.UserId,
				user.Email,
				user.FullName,
				user.Active,
				user.CreatedAt,
			)
		if err := s.doExec(tx, insertBuilder); err != nil {
			return fmt.Errorf("cannot do transaction exec: %v", err)
		}
		if err := tx.Commit(s.ctx); err != nil {
			return fmt.Errorf("cannot commit transaction: %v", err)
		}
		return nil

	}, &utils.Option{
		RetryCount:   txRetryCount,
		WaitInterval: txWaitInterval,
	})
}

func (s *storage) GetUser(userId, userEmail string) (*ServiceUser, bool, error) {
	builder := sq.
		Select(
			`user_id`,
			`email`,
			`password_hash`,
			`full_name`,
			`active`,
			`created_at`,
		).
		From(`stock_service_user`)

	where := sq.Eq{}
	if userId != "" {
		where[`user_id`] = userId
	}
	if userEmail != "" {
		where[`email`] = userEmail
	}
	builder = builder.Where(where)

	user := &ServiceUser{}

	found, err := s.doQuery(nil, builder,
		&user.UserId,
		&user.Email,
		&user.FullName,
		&user.Active,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, false, fmt.Errorf("cannot do query: %v", err)
	}
	if !found {
		return nil, false, nil
	}
	return user, true, nil
}

func (s *storage) doExec(tx pgx.Tx, builder queryBuilder) error {
	query, args := mustBuildQuery(builder)
	var (
		execFunc postgres.ExecFunc
	)
	if tx != nil {
		execFunc = tx.Exec
	} else {
		execFunc = s.client.Exec
	}
	if _, err := execFunc(s.ctx, query, args...); err != nil {
		return fmt.Errorf("cannot do exec: %v", err)
	}
	return nil
}

func (s *storage) doQuery(tx pgx.Tx, builder queryBuilder, fields ...any) (bool, error) {
	query, args := mustBuildQuery(builder)
	var (
		queryFunc postgres.QueryFunc
	)
	if tx != nil {
		queryFunc = tx.Query
	} else {
		queryFunc = s.client.Query
	}
	rows, err := queryFunc(s.ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("cannot do query: %v", err)
	}
	return scanFirstQueriedRow(rows, fields)
}

func mustBuildQuery(builder queryBuilder) (string, []any) {
	query, args, err := builder.ToSql()
	if err != nil {
		log.Fatalf("build insert query failed: %v", err)
	}
	return query, args
}

func scanFirstQueriedRow(rows pgx.Rows, fields []any) (bool, error) {
	var hasScannedRow bool
	if rows.Next() {
		if err := rows.Scan(fields...); err != nil {
			return false, fmt.Errorf("cannot scan queried row: %v", err)
		}
		hasScannedRow = true
	}
	return hasScannedRow, nil
}

func (s *storage) PutToken(token *RefreshToken) error {
	builder := sq.
		Insert(`refresh_token`).
		Columns(
			`token_id`,
			`active`,
		).
		Values(
			token.TokenId,
			token.Active,
		)
	if err := s.doExec(nil, builder); err != nil {
		return fmt.Errorf("cannot do exec: %v", err)
	}
	return nil
}

func (s *storage) GetToken(tokenId string) (*RefreshToken, bool, error) {
	builder := sq.
		Select(
			`token_id`,
			`active`,
			`user_id`,
		).
		From(`refresh_token`).
		Where(sq.Eq{
			`token_id`: tokenId,
		})
	token := &RefreshToken{}

	found, err := s.doQuery(nil, builder,
		&token.TokenId,
		&token.Active,
		&token.UserId,
	)
	if err != nil {
		return nil, false, fmt.Errorf("cannot do query: %v", err)
	}
	if !found {
		return nil, false, nil
	}

	return token, true, nil
}

func (s *storage) UpdateToken(token *RefreshToken) error {
	builder := sq.
		Update(`refresh_token`).
		Set(`active`, token.Active)

	if err := s.doExec(nil, builder); err != nil {
		return fmt.Errorf("cannot do exec: %v", err)
	}
	return nil
}
