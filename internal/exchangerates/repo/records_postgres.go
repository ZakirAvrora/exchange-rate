package repo

import (
	"context"
	"errors"
	"time"

	"github.com/ZakirAvrora/exchange-rate/internal/exchangerates"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type recordsRepository struct {
	connPool *pgxpool.Pool
}

func NewRecordsRepository(connPool *pgxpool.Pool) (*recordsRepository, error) {

	if connPool == nil {
		return nil, errors.New("provided connPool handle is nil")
	}

	return &recordsRepository{connPool}, nil
}

func (r *recordsRepository) Insert(ctx context.Context, record *exchangerates.Record) error {
	current_time := time.Now()
	rate := 0

	_, err := r.connPool.Exec(ctx, "INSERT INTO records (identifier, base, secondary, rate, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)",
		record.Identifier, record.Base, record.Secondary, rate, record.Status, current_time, current_time)

	if err != nil {
		return err
	}

	return nil
}
func (r *recordsRepository) FetchByIdentifier(ctx context.Context, identifier string) (*exchangerates.Record, error) {
	record := exchangerates.Record{}

	err := r.connPool.QueryRow(ctx, "SELECT * FROM records WHERE identifier = $1 LIMIT 1", identifier).Scan(
		&record.Id, &record.Identifier, &record.Base, &record.Secondary, &record.Rate,
		&record.Status, &record.Created_At, &record.Updated_At)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, exchangerates.ErrNoRecord
		}
		return nil, err
	}

	return &record, nil
}

func (r *recordsRepository) FetchLatest(ctx context.Context, base string, secondary string) (*exchangerates.Record, error) {
	record := exchangerates.Record{}

	err := r.connPool.QueryRow(ctx, "SELECT * FROM records WHERE base = $1 AND secondary = $2 AND STATUS = $3 ORDER BY ID DESC LIMIT 1",
		base, secondary, exchangerates.StatusUpdated).Scan(
		&record.Id, &record.Identifier, &record.Base, &record.Secondary, &record.Rate,
		&record.Status, &record.Created_At, &record.Updated_At)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, exchangerates.ErrNoRecord
		}
		return nil, err
	}

	return &record, nil
}

func (r *recordsRepository) Update(ctx context.Context, identifier string, rate float64) error {
	current_time := time.Now()
	_, err := r.connPool.Exec(ctx, "UPDATE records SET rate = $1, status = $2, updated_at = $3 where identifier = $4",
		rate, exchangerates.StatusUpdated, current_time, identifier)

	return err
}
