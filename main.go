package main

import (
    "fmt"
    "context"
    "time"
    "github.com/joho/godotenv"
    "log"
    "os"
    "net/http"
    "encoding/json"
    "database/sql"
    "github.com/Greeshmanth1909/blogAggregator/internal/database"
    "github.com/google/uuid"
    "strings"
)
import _ "github.com/lib/pq"

func main() {
    err := godotenv.Load()
    if err != nil {
    log.Fatal("Error loading .env file")
    }

    port := os.Getenv("PORT")
    dbURL := os.Getenv("URL")
    fmt.Println(port)
    // open connection to database
    db, error := sql.Open("postgres", dbURL)
    if error != nil {
        log.Fatal("Error establishing a connection to the database")
    }

    dbQueries := database.New(db)
    apiConf.DB = dbQueries

    mux := http.NewServeMux()

    var server http.Server
    server.Addr = "localhost:" + port
    server.Handler = mux

    mux.HandleFunc("GET /v1/healthz", healthHandler)
    mux.HandleFunc("GET /v1/err", errHandler)
    mux.HandleFunc("POST /v1/users", usersHandler)
    mux.HandleFunc("GET /v1/users", getUsersHandler)

    server.ListenAndServe()
}

type apiConfig struct {
	DB *database.Queries
}
var apiConf apiConfig

type user struct {
    Id uuid.UUID `json:"id"`
    Created_at time.Time `json:"created_at"`
    Updated_at time.Time `json:"updated_at"`
    Name string `json:"name"`
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    dat, err := json.Marshal(payload) 
    if err != nil {
        fmt.Println("something up with marshalling payload")
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(dat)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
    body := struct{
        Error string `json:"error"`
    }{msg}
    respondWithJSON(w, code, body)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    body := struct{
        Status string `json:"status"`
    }{"OK"}
    respondWithJSON(w, 200, body)
}

func errHandler(w http.ResponseWriter, r *http.Request) {
    respondWithError(w, 500, "Internal Server Error")
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    type body struct {
        Name string `json:name`
    }
    var req body
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&req)

    var user database.CreateUserParams
    // create user in db
    id := uuid.New()
    user.ID = id
    user.CreatedAt = time.Now()
    user.UpdatedAt = time.Now()
    user.Name = req.Name
    ctx := context.Background()
    usr, err := apiConf.DB.CreateUser(ctx, user)
    if err != nil {
        respondWithError(w, 500, "Internal Server Error")
    } 
    respondWithJSON(w, 200, usr)
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
    api_key := r.Header.Get("Authorization")
    api_key = strings.TrimPrefix(api_key, "ApiKey ")
    ctx := context.Background()
    user, err := apiConf.DB.GetUserByApi(ctx, api_key)
    if err != nil {
        respondWithError(w, 500, "Internal server Error")
    }
    respondWithJSON(w, 200, user)
}
