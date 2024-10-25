package gorm

import (
	"context"
	"fmt"
	"github.com/e154/smart-home/db"
	"gorm.io/gorm"
)

func InjectTransaction(ctx context.Context, tr *gorm.DB) context.Context {
	return context.WithValue(ctx, db.GormTransaction, tr)
}

type TransactionManger struct {
	db *gorm.DB
}

func NewTransactionManger(db *gorm.DB) *TransactionManger {
	return &TransactionManger{db: db}
}

func (m *TransactionManger) Do(ctx context.Context, fn func(context.Context) error) (doErr error) {
	tr := m.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tr.Rollback()
			if doErr == nil {
				doErr = fmt.Errorf("Recover")
			}
			return
		}

		if doErr != nil {
			tr.Rollback()
			return
		}
		tr.Commit()
	}()

	return fn(InjectTransaction(ctx, tr))
}
