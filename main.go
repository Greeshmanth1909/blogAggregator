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
    mux.Handle("GET /v1/users", authMiddleWare(http.HandlerFunc(getUsersHandler)))
    mux.Handle("POST /v1/feeds", authMiddleWare(http.HandlerFunc(createFeedHandler)))
    mux.HandleFunc("GET /v1/feeds", getFeedsHandler)
    mux.Handle("POST /v1/feed_follows", authMiddleWare(http.HandlerFunc(createFeedFollow)))

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

func authMiddleWare(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        api_key := r.Header.Get("Authorization")
        api_key = strings.TrimPrefix(api_key, "ApiKey ")
        // make sure the thing exists in the db
        ctx := context.Background()
        user, err := apiConf.DB.GetUserByApi(ctx, api_key)
        ctx = context.WithValue(ctx, "user", user)
        r = r.WithContext(ctx)
        if err != nil {
            respondWithJSON(w, 404, "Invalid api key")
            return
        }
        handler.ServeHTTP(w, r)
    })
}

func createFeedHandler(w http.ResponseWriter, r *http.Request) {
    type feedStruct struct {
        Name string `json:"name"`
        Url string `json:"url"`
    }
    var body feedStruct
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&body)
    // add name and url with correnponding user data
    var feed database.CreateFeedParams
    feed.ID = uuid.New()
    feed.CreatedAt = time.Now()
    feed.UpdatedAt = time.Now()
    feed.Name = body.Name
    feed.Url = body.Url
    user := r.Context().Value("user").(database.User)
    feed.UserID = user.ID
    ctx := context.Background()
    fd, err := apiConf.DB.CreateFeed(ctx, feed)
    if err != nil {
        respondWithJSON(w, 500, "Couldn't create feed")
        return
    }
    respondWithJSON(w, 200, fd)
}

func getFeedsHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    dat, err := apiConf.DB.GetAllFeeds(ctx)
    if err != nil {
        respondWithError(w, 500, "Internal server error: couldn't query database")
        return
    }
    respondWithJSON(w, 200, dat)
}

func createFeedFollow(w http.ResponseWriter, r *http.Request) {
    type body struct{
        Feed_id string `json:"feed_id"`
    }
    var reqBody body
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&reqBody)
    user := r.Context().Value("user").(database.User)

    var params database.AddFeedFollowParams
    params.ID = uuid.New()
    params.FeedID, _ = uuid.Parse(reqBody.Feed_id)
    params.UserID = user.ID
    params.CreatedAt = time.Now()
    params.UpdatedAt = time.Now()
    
    ctx := context.Background()
    res, err := apiConf.DB.AddFeedFollow(ctx, params)
    if err != nil {
        respondWithError(w, 500, "Server Error: couldn't update database")
        return
    }
    respondWithJSON(w, 200, res)

}
