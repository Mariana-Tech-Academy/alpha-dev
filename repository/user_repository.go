package repository

import (
	"gorm.io/gorm"
)

type UserRepository interface {

}

type PostgresUserRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepo{db: db}
}
