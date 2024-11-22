package database

import (
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

type DBItem struct {
	ID    int64                  `json:"id"`
	Name  string                 `json:"name"`
	Value map[string]interface{} `json:"value"`
}

type DBAdapter interface {
	InitialzeDB() error
	FinishDB() error
	Insert(item *DBItem, userHash string) (int64, error)
	Update(id int64, item *DBItem, userHash string) error
	Delete(id int64, userHash string) error
	Get(id int64, userHash string) (*DBItem, error)
	GetAll(userHash string) ([]*DBItem, error)
}

func NewDBAdapter(dbType string, cfg *DBConfig) (DBAdapter, error) {
	switch dbType {
	case "postgres":
		return NewPostgresAdapter(cfg), nil
	default:
		return nil, fmt.Errorf("unknown db type %s", dbType)
	}
}
