package api

import (
	"auth_service/database"
	"auth_service/models"
	"encoding/json"
	"log"
	"net/http"
	json_utils "utils/json"
	jwt_utils "utils/jwt"
	redis_utils "utils/redis"

	"golang.org/x/crypto/bcrypt"
)

type API struct {
	db        database.DBAdapter
	redis_cli *redis_utils.RedisClient
}

func NewApi(db database.DBAdapter, redis_cli *redis_utils.RedisClient) *API {
	return &API{db, redis_cli}
}

func (api *API) HandlerAuthorize(w http.ResponseWriter, req *http.Request) {
	log.Println("HandlerAuthorize called")
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

	var input models.Credentials
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	user, err := api.db.Get(input.Username)
	if err != nil {
		http.Error(w, "Error get user from db", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		http.Error(w, "Error get user from db", http.StatusUnauthorized)
		return
	}

	token, err := jwt_utils.GenerateJWT(user.Username)
	if err != nil {
		log.Printf("GenerateJWT: Error: %v", err)
		http.Error(w, "Token generation failed", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (api *API) HandlerRegister(w http.ResponseWriter, req *http.Request) {
	log.Println("HandlerRegister called")
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

	var input models.Credentials
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error generating hash grom password", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Username: input.Username,
		Password: string(hashedPassword)}

	id, err := api.db.Insert(user)
	if err != nil {
		http.Error(w, "Error when create user", http.StatusInternalServerError)
		return
	}
	user.ID = id

	json_utils.SendJSONResponse(w, http.StatusOK, map[string]int64{"id": id})
	log.Printf("Created user: %+v", user)
}

func (api *API) HandlerLogout(w http.ResponseWriter, req *http.Request) {
	log.Println("HandlerLogout called")

	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

	token, err := jwt_utils.GetJwtTokenFromHeader(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	claims, err := jwt_utils.GetClaimsFromContext(req)
	if err != nil {
		log.Printf("handleUpdate: Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = api.redis_cli.RevokeToken(token, claims.ExpiresAt.Time)
	if err != nil {
		log.Printf("Logout error: %+v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json_utils.SendJSONResponse(w, http.StatusOK, nil)
	log.Printf("Logout token: %+v", token)
}
