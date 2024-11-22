package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

const PSQL_ERR_DB_ALREADY_EXISTS = "42P04"

type PostgresDB struct {
	db  *sql.DB
	cfg *DBConfig
}

func NewPostgresAdapter(cfg *DBConfig) *PostgresDB {
	return &PostgresDB{
		db:  nil,
		cfg: cfg,
	}
}

func (p *PostgresDB) InitialzeDB() error {
	db, err := p.сonnectAndCreateDB(p.cfg)
	if err != nil {
		return err
	}

	p.db = db
	return p.initializeTables()
}

func (p *PostgresDB) сonnectAndCreateDB(cfg *DBConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Проверяем, существует ли целевая база данных
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))
	if err != nil {
		// Проверяем, является ли ошибка "база данных уже существует"
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == PSQL_ERR_DB_ALREADY_EXISTS {
			// Код ошибки "42P04" соответствует "database already exists"
			// Игнорируем эту ошибку
		} else {
			return nil, err
		}
	}

	db.Close()

	// Подключаемся к созданной (или существующей) базе данных
	connStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	return sql.Open("postgres", connStr)
}

func (p *PostgresDB) initializeTables() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
				id SERIAL PRIMARY KEY,
				username VARCHAR(50) UNIQUE NOT NULL,
				password VARCHAR(512) NOT NULL
		);`
	_, err := p.db.Exec(query)
	return err
}

func (p *PostgresDB) FinishDB() error {
	return p.db.Close()
}

func (p *PostgresDB) Insert(item *DBItem) (int64, error) {
	var err error

	query := `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id`
	var id int64

	err = p.db.QueryRow(query, item.Username, item.Password).Scan(&id)
	return id, err
}

func (p *PostgresDB) Get(name string) (*DBItem, error) {
	query := `SELECT id, password FROM users WHERE username=$1`
	v := &DBItem{}

	err := p.db.QueryRow(query, name).Scan(&v.ID, &v.Password)
	if err != nil {
		return nil, errors.New("failed to select user from db: " + err.Error())
	}
	v.Username = name

	return v, nil
}

func (p *PostgresDB) GetAll() ([]*DBItem, error) {
	query := `SELECT id, username FROM users`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Обязательно закрываем rows после обработки

	var items []*DBItem
	// Итерируемся по результатам
	for rows.Next() {
		v := &DBItem{}
		// Сканируем текущую строку в структуру DBItem
		if err := rows.Scan(&v.ID, &v.Username); err != nil {
			return nil, err
		}

		items = append(items, v)
	}

	// Проверка на ошибки, возникшие во время итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (p *PostgresDB) Delete(name string) error {
	query := `DELETE FROM users WHERE username=$1`
	result, err := p.db.Exec(query, name)
	if err != nil {
		return err
	}

	// Проверяем, было ли удалено хотя бы одно значение
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

func (p *PostgresDB) Update(id int64, item *DBItem) error {
	// var jsonValue interface{} = nil
	// if item.Value != nil {
	// 	var err error
	// 	jsonValue, err = json.Marshal(item.Value)
	// 	if err != nil {
	// 		return fmt.Errorf("error marshalling json: %v", err)
	// 	}
	// }

	// query := `UPDATE items
	//            SET
	//             name = COALESCE(NULLIF($1, ''), name),
	//             value = COALESCE($2, value)
	//            WHERE id=$3`
	// result, err := p.db.Exec(query, item.Name, jsonValue, id)
	// if err != nil {
	// 	return err
	// }

	// // Проверяем, было ли обновленно хотя бы одно значение
	// rowsAffected, err := result.RowsAffected()
	// if err != nil {
	// 	return err
	// }

	// if rowsAffected == 0 {
	// 	return errors.New("no rows affected")
	// }

	return nil
}
