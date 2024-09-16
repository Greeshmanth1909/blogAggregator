package main

import (
    "fmt"
    "github.com/joho/godotenv"
    "log"
    "os"
    "net/http"
    "encoding/json"
)
func main() {
    err := godotenv.Load()
    if err != nil {
    log.Fatal("Error loading .env file")
    }

    port := os.Getenv("PORT")
    fmt.Println(port)

    mux := http.NewServeMux()

    var server http.Server
    server.Addr = "localhost:" + port
    server.Handler = mux

    mux.HandleFunc("GET /v1/healthz", healthHandler)
    mux.HandleFunc("GET /v1/err", errHandler)

    server.ListenAndServe()
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
