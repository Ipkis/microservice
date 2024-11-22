package main

import (
	"fmt"
	"log"
	"mock_service/api"
	"mock_service/database"
	"net/http"
	"os"
	jwt_utils "utils/jwt"
	redis_utils "utils/redis"
)

func main() {
	config, err := ReadConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Loaded Config: %+v", config)

	db_adapter, err := database.NewDBAdapter("postgres", &config.DBConf)
	if err != nil {
		log.Fatal(err)
	}
	defer db_adapter.FinishDB()
	log.Printf("Created database adapter")

	log.Printf("Initialze database...")
	err = db_adapter.InitialzeDB()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Init Redis client")
	r_addr := os.Getenv("REDIS_HOST")
	r_port := os.Getenv("REDIS_PORT")
	redis_cli := redis_utils.InitRedisClient(r_addr+":"+r_port, "")

	secret := os.Getenv("JWT_SECRET")
	jwt_utils.InitJwtSecret(secret)

	api := api.NewApi(db_adapter)

	mux := http.NewServeMux()
	mux.Handle("/data", jwt_utils.JwtMiddleware(http.HandlerFunc(api.HandlerData), redis_cli))
	mux.Handle("/list", jwt_utils.JwtMiddleware(http.HandlerFunc(api.HandlerList), redis_cli))

	port := fmt.Sprintf("%d", config.Port)
	log.Printf("Starting server on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
