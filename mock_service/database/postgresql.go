package database

import (
	"database/sql"
	"encoding/json"
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
    CREATE TABLE IF NOT EXISTS _items (
        id BIGSERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        value JSONB NOT NULL
    );

		CREATE TABLE IF NOT EXISTS _user2items (
				id BIGSERIAL PRIMARY KEY,
				userhash TEXT NOT NULL,
				item_id BIGINT NOT NULL REFERENCES _items(id) ON DELETE CASCADE
		);

		CREATE OR REPLACE VIEW items AS
			SELECT 
					u.userhash AS userhash,
					i.id AS id,
					i.name AS name,
					i.value AS value
			FROM _user2items u
			JOIN _items i ON u.item_id = i.id;

		CREATE OR REPLACE FUNCTION insert_into_items(name TEXT, value JSONB, userhash TEXT)
		RETURNS BIGINT AS $$
		DECLARE
				new_item_id BIGINT;
		BEGIN
				INSERT INTO _items (name, value)
				VALUES (name, value)
				RETURNING id INTO new_item_id;

				INSERT INTO _user2items (userhash, item_id)
				VALUES (userhash, new_item_id);

				RETURN new_item_id;
		END;
		$$ LANGUAGE plpgsql;

		CREATE OR REPLACE RULE delete_user_items AS
			ON DELETE TO items
			DO INSTEAD (
					DELETE FROM _items 
						WHERE id IN (
							SELECT item_id FROM _user2items WHERE item_id = OLD.id AND userhash = OLD.userhash
						);
		);
		CREATE OR REPLACE RULE update_user_items AS
			ON UPDATE TO items
			DO INSTEAD (
				UPDATE _items
				SET
					name = COALESCE(NULLIF(NEW.name, ''), OLD.name), 
					value = COALESCE(NEW.value, OLD.value) 
				WHERE id IN (
					SELECT item_id FROM _user2items WHERE item_id = OLD.id AND userhash = OLD.userhash
				);
		);
		`
	_, err := p.db.Exec(query)
	return err
}

func (p *PostgresDB) FinishDB() error {
	return p.db.Close()
}

func (p *PostgresDB) Insert(item *DBItem, userHash string) (int64, error) {
	var jsonValue interface{} = nil
	var err error
	if item.Value != nil {
		jsonValue, err = json.Marshal(item.Value)
		if err != nil {
			return 0, fmt.Errorf("error marshalling json: %v", err)
		}
	}

	query := `select insert_into_items(NULLIF($1, ''), $2, $3);`
	var id int64
	err = p.db.QueryRow(query, item.Name, jsonValue, userHash).Scan(&id)
	return id, err
}

func (p *PostgresDB) Get(id int64, userHash string) (*DBItem, error) {
	query := `SELECT id, name, value FROM items WHERE id=$1 and userhash=$2`
	v := &DBItem{}
	var valueBytes []byte
	err := p.db.QueryRow(query, id, userHash).Scan(&v.ID, &v.Name, &valueBytes)
	if err != nil {
		return nil, errors.New("failed to select item from db: " + err.Error())
	}

	// Декодируем JSON-значение из valueBytes в map[string]interface{}
	if err := json.Unmarshal(valueBytes, &v.Value); err != nil {
		return nil, errors.New("failed to decode JSON field 'value': " + err.Error())
	}

	return v, nil
}

func (p *PostgresDB) GetAll(userHash string) ([]*DBItem, error) {
	query := `SELECT id, name, value FROM items WHERE userhash=$1 ORDER BY id`
	rows, err := p.db.Query(query, userHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Обязательно закрываем rows после обработки

	var items []*DBItem
	// Итерируемся по результатам
	for rows.Next() {
		v := &DBItem{}
		var valueBytes []byte // временная переменная для хранения []byte из jsonb
		// Сканируем текущую строку в структуру DBItem
		if err := rows.Scan(&v.ID, &v.Name, &valueBytes); err != nil {
			return nil, err
		}

		// Декодируем JSON-значение из valueBytes в map[string]interface{}
		if err := json.Unmarshal(valueBytes, &v.Value); err != nil {
			return nil, errors.New("failed to decode JSON field 'value': " + err.Error())
		}

		items = append(items, v)
	}

	// Проверка на ошибки, возникшие во время итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (p *PostgresDB) Delete(id int64, userHash string) error {
	query := `DELETE FROM items WHERE id=$1 AND userhash=$2`
	result, err := p.db.Exec(query, id, userHash)
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

func (p *PostgresDB) Update(id int64, item *DBItem, userHash string) error {
	var jsonValue interface{} = nil
	if item.Value != nil {
		var err error
		jsonValue, err = json.Marshal(item.Value)
		if err != nil {
			return fmt.Errorf("error marshalling json: %v", err)
		}
	}

	query := `UPDATE items 
             SET
              name = COALESCE(NULLIF($1, ''), name), 
              value = COALESCE($2, value) 
             WHERE id=$3 AND userhash=$4`
	result, err := p.db.Exec(query, item.Name, jsonValue, id, userHash)
	if err != nil {
		return err
	}

	// Проверяем, было ли обновленно хотя бы одно значение
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}
