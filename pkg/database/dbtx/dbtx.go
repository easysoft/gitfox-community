package dbtx

import (
	"context"

	"gorm.io/gorm"
)

// ctxKeyTx is the context key for the transaction.
type ctxKeyTx struct{}

// TxRunner provides transaction management.
type TxRunner interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

// TxRunnerImpl implements TxRunner using GORM.
type TxRunnerImpl struct {
	db *gorm.DB
}

// NewTxRunner creates a new TxRunner.
func NewTxRunner(db *gorm.DB) *TxRunnerImpl {
	return &TxRunnerImpl{db: db}
}

// WithTx executes a function within a transaction.
func (r *TxRunnerImpl) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Store the transaction in context
		txCtx := context.WithValue(ctx, ctxKeyTx{}, tx)
		return fn(txCtx)
	})
}

// GetOrmAccessor returns the *gorm.DB from context if in a transaction,
// otherwise returns the provided db with context.
func GetOrmAccessor(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(ctxKeyTx{}).(*gorm.DB); ok {
		return tx
	}
	return db.WithContext(ctx)
}
