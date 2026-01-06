package branch

import (
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

type Store struct {
	db *gorm.DB
}

func NewBranchStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}
