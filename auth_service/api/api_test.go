package api

import (
	"auth_service/models"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type MockDB struct {
	Users map[string]*models.User
}

func (db *MockDB) InitialzeDB() error {
	return nil
}

func (db *MockDB) FinishDB() error {
	return nil
}

func (db *MockDB) Get(username string) (*models.User, error) {
	if user, ok := db.Users[username]; ok {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (db *MockDB) Insert(user *models.User) (int64, error) {
	if _, ok := db.Users[user.Username]; ok {
		return 0, errors.New("user already exists")
	}
	id := int64(len(db.Users) + 1)
	user.ID = id
	db.Users[user.Username] = user
	return id, nil
}

func (db *MockDB) Update(id int64, item *models.User) error {
	return nil
}

func (db *MockDB) Delete(username string) error {
	return nil
}

func (db *MockDB) GetAll() ([]*models.User, error) {
	return nil, nil
}

func TestHandlerAuthorize(t *testing.T) {
	log.SetOutput(io.Discard)

	mockDB := &MockDB{
		Users: map[string]*models.User{
			"testuser": {
				Username: "testuser",
				Password: func() string {
					hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
					return string(hash)
				}(),
			},
		},
	}
	api := NewApi(mockDB, nil)

	t.Run("Valid Credentials", func(t *testing.T) {
		reqBody, _ := json.Marshal(models.Credentials{Username: "testuser", Password: "password123"})
		req := httptest.NewRequest(http.MethodPost, "/authorize", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		api.HandlerAuthorize(w, req)
		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.NotNil(t, response["token"])
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		reqBody, _ := json.Marshal(models.Credentials{Username: "testuser", Password: "1"})
		req := httptest.NewRequest(http.MethodPost, "/authorize", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		api.HandlerAuthorize(w, req)
		resp := w.Result()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

}

func TestHandlerRegister(t *testing.T) {
	log.SetOutput(io.Discard)

	mockDB := &MockDB{Users: make(map[string]*models.User)}
	api := NewApi(mockDB, nil)

	t.Run("New user", func(t *testing.T) {
		reqBody, _ := json.Marshal(models.Credentials{Username: "newuser", Password: "securepassword"})
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		api.HandlerRegister(w, req)
		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]int64
		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, int64(1), response["id"])
		assert.NotNil(t, mockDB.Users["newuser"])
	})

	t.Run("Existing user", func(t *testing.T) {
		reqBody, _ := json.Marshal(models.Credentials{Username: "newuser", Password: "securepassword"})
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		api.HandlerRegister(w, req)
		resp := w.Result()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

}

func TestHandlerLogout(t *testing.T) {
	// api := NewApi(nil)

	// req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	// w := httptest.NewRecorder()

	// api.HandlerLogout(w, req)
	// resp := w.Result()
	// assert.Equal(t, http.StatusOK, resp.StatusCode)
	// assert.True(t, mockRedis.RevokedTokens["mocked_token"])
}
