package storage

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	sq "github.com/Masterminds/squirrel"
	"github.com/UshakovN/stock-predictor-service/postgres"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

const counterInc = 1

type queryBuilder interface {
	ToSql() (string, []any, error)
}

type Storage interface {
	PutStoredMedia(storedMedia *StoredMedia) error
	GetStoredMedia(storedMediaId string) (*StoredMedia, bool, error)
	GetStoredMediaBatch(storedMediaIds []string) ([]*StoredMedia, error)
}

type storage struct {
	ctx     context.Context
	client  postgres.Client
	counter atomic.Uint64
}

func NewStorage(ctx context.Context, config *postgres.Config) (Storage, error) {
	client, err := postgres.NewClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("cannot create new postgres client: %v", err)
	}
	return &storage{
		ctx:     ctx,
		client:  client,
		counter: atomic.Uint64{},
	}, nil
}

func (s *storage) PutStoredMedia(storedMedia *StoredMedia) error {
	builder := sq.Insert(`stored_media`).
		Columns(
			`stored_media_id`,
			`formed_url`,
			`created_by`,
			`created_at`,
		).
		Values(
			storedMedia.StoredMediaId,
			storedMedia.FormedURL,
			storedMedia.CreatedBy,
			storedMedia.CreatedAt,
		).
		Suffix(`ON CONFLICT (stored_media_id) DO NOTHING`).
		PlaceholderFormat(sq.Dollar)

	if err := s.doPutQuery(builder); err != nil {
		return err
	}
	log.Infof("put stored media with id '%s'. total: %d",
		storedMedia.StoredMediaId, s.counter.Add(counterInc))

	return nil
}

func (s *storage) GetStoredMedia(storedMediaId string) (*StoredMedia, bool, error) {
	builder := sq.
		Select(
			`stored_media_id`,
			`formed_url`,
			`created_by`,
			`created_at`,
		).
		From(`stored_media`).
		Where(sq.Eq{
			`stored_media_id`: storedMediaId,
		}).
		PlaceholderFormat(sq.Dollar)

	storedMedia := &StoredMedia{}

	found, err := s.doGetQuery(builder,
		&storedMedia.StoredMediaId,
		&storedMedia.FormedURL,
		&storedMedia.CreatedBy,
		&storedMedia.CreatedAt,
	)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}

	return storedMedia, true, nil
}

func (s *storage) GetStoredMediaBatch(storedMediaIds []string) ([]*StoredMedia, error) {
	_ = sq.
		Select(
			`stored_media_id`,
			`formed_url`,
			`created_by`,
			`created_at`,
		).
		From(`stored_media`).
		Where(sq.Eq{
			`stored_media_id`: storedMediaIds,
		})

	// TODO: complete this

	return nil, nil
}

func (s *storage) doPutQuery(builder queryBuilder) error {
	query, args := mustBuildQuery(builder)
	if _, err := s.client.Exec(s.ctx, query, args...); err != nil {
		return fmt.Errorf("cannot do exec: %v", err)
	}
	return nil
}

func mustBuildQuery(builder queryBuilder) (string, []any) {
	query, args, err := builder.ToSql()
	if err != nil {
		log.Fatalf("build insert query failed: %v", err)
	}
	return query, args
}

func (s *storage) doGetQuery(builder queryBuilder, fields ...any) (bool, error) {
	query, args := mustBuildQuery(builder)
	rows, err := s.client.Query(s.ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("cannot do query: %v", err)
	}
	return scanFirstQueriedRow(rows, fields)
}

func scanFirstQueriedRow(rows pgx.Rows, fields []any) (bool, error) {
	var (
		hasRows bool
		err     error
	)
	once := sync.Once{}
	for rows.Next() {
		once.Do(func() {
			err = rows.Scan(fields...)
			hasRows = true
		})
	}
	return hasRows, err
}
