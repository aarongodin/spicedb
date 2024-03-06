package util

import (
	"context"
	"database/sql"
	"errors"

	log "github.com/authzed/spicedb/internal/logging"
)

var (
	NoopTxCleanupFunc = func() error { return nil }
)

type TxUserFunc = func(*sql.Tx) error
type TxCleanupFunc func() error
type TxFactory func(context.Context) (*sql.Tx, TxCleanupFunc, error)

// WithTx is a helper for managing a transaction with user code called in a callback
func WithTx(ctx context.Context, db *sql.DB, txOptions *sql.TxOptions, fn TxUserFunc) error {
	tx, err := db.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		rerr := tx.Rollback()
		if rerr != nil {
			return errors.Join(err, rerr)
		}

		return err
	}

	return tx.Commit()
}

func (txf TxFactory) WithTx(ctx context.Context, fn TxUserFunc) error {
	tx, txCleanup, err := txf(ctx)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		rerr := txCleanup()
		if rerr != nil {
			log.Ctx(ctx).Err(rerr).Msg("datastore error")
			return errors.Join(err, rerr)
		}

		return err
	}
	return nil
}

// NewTxFactory creates a tx factory that gives back a new transaction on every call
func NewTxFactory(db *sql.DB) TxFactory {
	return func(ctx context.Context) (*sql.Tx, TxCleanupFunc, error) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return nil, nil, err
		}

		return tx, tx.Rollback, nil
	}
}

// NewTxFactory creates a tx factory that always returns an already started transaction
func NewTxFactoryWithInstance(tx *sql.Tx) TxFactory {
	return func(context.Context) (*sql.Tx, TxCleanupFunc, error) {
		return tx, NoopTxCleanupFunc, nil
	}
}
