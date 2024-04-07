package exchangerates

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
)

type Recorder struct {
	repo  RecordsRepo
	queue chan Record
}

func NewService(repo RecordsRepo) *Recorder {
	return &Recorder{
		repo:  repo,
		queue: make(chan Record, 5),
	}
}

func (r *Recorder) Queue() <-chan Record {
	return r.queue
}

func (r *Recorder) Refresh(ctx context.Context, base string, secondary string) (string, error) {
	base = strings.ToUpper(base)
	secondary = strings.ToUpper(secondary)

	if err := validPair(base, secondary); err != nil {
		return "", err
	}

	v4, err := uuid.NewRandom()
	if err != nil {
		log.Fatal("cannot generate unique identifier")
	}
	identifier := v4.String()

	record := &Record{
		Identifier: identifier,
		Base:       base,
		Secondary:  secondary,
		Status:     StatusCreated,
	}

	if err := r.repo.Insert(ctx, record); err != nil {
		return "", fmt.Errorf("refresh request is failed: %w", err)
	}

	r.queue <- *record

	return identifier, nil
}

func (r *Recorder) FetchByIdentifier(ctx context.Context, identifier string) (*Record, error) {
	record, err := r.repo.FetchByIdentifier(ctx, strings.TrimSpace(identifier))
	if err != nil {
		return nil, fmt.Errorf("fetching exchangerate by identifier is failed: %w", err)
	}
	return record, nil
}

func (r *Recorder) FetchLatest(ctx context.Context, base string, secondary string) (*Record, error) {
	base = strings.ToUpper(base)
	secondary = strings.ToUpper(secondary)

	if err := validPair(base, secondary); err != nil {
		return nil, err
	}

	record, err := r.repo.FetchLatest(ctx, base, secondary)
	if err != nil {
		return nil, fmt.Errorf("fetching latest exchangerate is failed: %w", err)
	}
	return record, nil
}

func (r *Recorder) ShiftUpdated(ctx context.Context, identifeir string, rate float64) error {
	err := r.repo.ShiftUpdated(ctx, identifeir, rate)
	return fmt.Errorf("shifting status to updated error: %w", err)
}

func (r *Recorder) ShiftFailed(ctx context.Context, identifeir string) error {
	err := r.repo.ShiftFailed(ctx, identifeir)
	return fmt.Errorf("shifting status to failed error: %w", err)
}

func validPair(base, secondary string) error {
	if !validBaseCurrency(base) {
		return ErrNotSupportedBaseCurrency
	}

	if !validSecondaryCurrency(secondary) {
		return ErrNotSupportedSecondaryCurrency
	}

	return nil
}
