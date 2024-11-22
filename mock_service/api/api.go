package api

import (
	"fmt"
	"log"
	"mock_service/database"
	"net/http"
	"strconv"

	json_utils "utils/json"
	jwt_utils "utils/jwt"
)

type API struct {
	db database.DBAdapter
}

func NewApi(db database.DBAdapter) *API {
	return &API{db}
}

func (api *API) HandlerData(w http.ResponseWriter, req *http.Request) {
	log.Printf("HandlerData called. METHOD: %s", req.Method)

	switch req.Method {
	case http.MethodGet:
		api.handleGet(w, req)
	case http.MethodPost:
		api.handleCreate(w, req)
	case http.MethodPut:
		api.handleUpdate(w, req)
	case http.MethodDelete:
		api.handleDelete(w, req)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (api *API) handleCreate(w http.ResponseWriter, req *http.Request) {
	var item database.DBItem
	if err := json_utils.DecodeJSONBody(req, &item); err != nil {
		return
	}

	claims, err := jwt_utils.GetClaimsFromContext(req)
	if err != nil {
		log.Printf("handleCreate: Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := api.db.Insert(&item, claims.Subject)
	if err != nil {
		log.Printf("handleCreate: Error: %v", err)
		http.Error(w, "Failed to insert item", http.StatusInternalServerError)
		return
	}

	json_utils.SendJSONResponse(w, http.StatusOK, map[string]int64{"id": id})
	log.Println("handleCreate: item created successfully")
}

func (api *API) handleUpdate(w http.ResponseWriter, req *http.Request) {
	id, err := parseIDParam(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var item database.DBItem
	if err := json_utils.DecodeJSONBody(req, &item); err != nil {
		return
	}

	claims, err := jwt_utils.GetClaimsFromContext(req)
	if err != nil {
		log.Printf("handleUpdate: Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = api.db.Update(id, &item, claims.Subject)
	if err != nil {
		if err.Error() == "no rows affected" {
			http.Error(w, "Item not found", http.StatusNotFound)
		} else {
			log.Printf("handleUpdate: Error: %v", err)
			http.Error(w, "Failed to update item", http.StatusInternalServerError)
		}
		return
	}

	json_utils.SendJSONResponse(w, http.StatusOK, map[string]string{"message": "Item updated successfully"})
	log.Println("handleUpdate: item updated successfully")
}

func (api *API) handleDelete(w http.ResponseWriter, req *http.Request) {
	id, err := parseIDParam(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	claims, err := jwt_utils.GetClaimsFromContext(req)
	if err != nil {
		log.Printf("handleUpdate: Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = api.db.Delete(id, claims.Subject)
	if err != nil {
		if err.Error() == "no rows affected" {
			http.Error(w, "Item not found", http.StatusNotFound)
		} else {
			log.Printf("handleDelete: Error: %v", err)
			http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		}
		return
	}

	json_utils.SendJSONResponse(w, http.StatusOK, map[string]string{"message": "Item deleted successfully"})
	log.Println("handleDelete: item deleted successfully")
}

func (api *API) handleGet(w http.ResponseWriter, req *http.Request) {
	log.Println("handleGet called")
	id, err := parseIDParam(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	claims, err := jwt_utils.GetClaimsFromContext(req)
	if err != nil {
		log.Printf("handleUpdate: Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	item, err := api.db.Get(id, claims.Subject)
	if err != nil {
		log.Printf("handleGet: Error in Get - %v", err)
		http.Error(w, "Failed to get item", http.StatusInternalServerError)
		return
	}

	json_utils.SendJSONResponse(w, http.StatusOK, item)
	log.Println("handleGet: finished successfully")
}

func (api *API) HandlerList(w http.ResponseWriter, req *http.Request) {
	log.Println("HandlerList called")

	claims, err := jwt_utils.GetClaimsFromContext(req)
	if err != nil {
		log.Printf("handleUpdate: Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	items, err := api.db.GetAll(claims.Subject)
	if err != nil {
		log.Printf("HandlerList: Error in GetAll - %v", err)
		http.Error(w, "Failed to get item list", http.StatusInternalServerError)
		return
	}

	json_utils.SendJSONResponse(w, http.StatusOK, items)
	log.Println("HandlerList: finished successfully")
}

// Вспомогательная функция для парсинга параметра ID из URL.
func parseIDParam(req *http.Request) (int64, error) {
	idParam := req.URL.Query().Get("id")
	if idParam == "" {
		return 0, fmt.Errorf("ID parameter is required")
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Printf("HandlerData: error parsing id - %v", err)
		return 0, fmt.Errorf("invalid ID")
	}
	return id, nil
}
