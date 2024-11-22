package database

import (
	"auth_service/models"
	"fmt"

	_ "github.com/lib/pq"
)

type DBConfig struct {
	DBHost     string `json:"db_host"`
	DBPort     int    `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`
}

type DBItem = models.User

type DBAdapter interface {
	InitialzeDB() error
	FinishDB() error
	Insert(item *DBItem) (int64, error)
	Update(id int64, item *DBItem) error
	Delete(name string) error
	Get(name string) (*DBItem, error)
	GetAll() ([]*DBItem, error)
}

func NewDBAdapter(dbType string, cfg *DBConfig) (DBAdapter, error) {
	switch dbType {
	case "postgres":
		return NewPostgresAdapter(cfg), nil
	default:
		return nil, fmt.Errorf("unknown db type %s", dbType)
	}
}
