package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"github.com/ablanchetMD/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	Db *database.Queries
	Platform string
	fileserverHits uint64
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || (len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/app/") {
			atomic.AddUint64(&cfg.fileserverHits, 1)
		}
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	atomic.StoreUint64(&cfg.fileserverHits, 0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Count reset to 0"))
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg := &apiConfig{}
	godotenv.Load(".env")
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		fmt.Println("Error fetching database: ", err)
		return
	}
	defer db.Close()
	dbQueries := database.New(db)
	cfg.Db = dbQueries
	cfg.Platform = os.Getenv("PLATFORM")

	mux := http.NewServeMux()
	fileserver := http.FileServer(http.Dir("."))
	mux.Handle("/app/", http.StripPrefix("/app", fileserver))

	handlerReadiness := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		handleCreateUser(cfg, w, r)
	})

	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		handleReset(cfg, w, r)
	})

	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		handleLogin(cfg, w, r)
	})
	// mux.HandleFunc("POST /api/login", db.handleLogin)handleLogin

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("./admin/index.html")
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(string(data), atomic.LoadUint64(&cfg.fileserverHits))))
	})

	 mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		handleCreateChirp(cfg, w, r)
	})

	 mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		handleGetChirps(cfg, w, r)
	})
	mux.HandleFunc("/api/chirps/{chirpID}", func(w http.ResponseWriter, r *http.Request) {
		handleGetChirp(cfg, w, r)
	})

	mux.HandleFunc("/api/reset", cfg.resetHandler)

	wrappedMux := middlewareLog(cfg.middlewareMetricsInc(mux))
	portString := "8080"
	srv := &http.Server{
		Addr:    ":" + portString,
		Handler: wrappedMux,
	}
	log.Printf("Server listening on port %s", portString)
	err = srv.ListenAndServe()
	if err != nil {
		fmt.Println("Error starting server: ", err)
		log.Fatal(err)
	}
}

// func main() {

// 	godotenv.Load(".env")

// 	portString := os.Getenv("PORT")
// 	if portString == "" {
// 		log.Fatal("PORT environment variable not set")
// 	}
// 	fmt.Printf("PORT: %s\n", portString)
// 	router := chi.NewRouter()

// 	router.Use(cors.Handler(cors.Options{
// 		AllowedOrigins:   []string{"https://*", "http://*"},
// 		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
// 		AllowedHeaders:   []string{"*"},
// 		ExposedHeaders:   []string{"Link"},
// 		AllowCredentials: false,
// 		MaxAge:           300, // Maximum value not ignored by any of major browsers

// 	}))

// 	v1Router := chi.NewRouter()

// 	v1Router.Get("/healthz", handlerReadiness)
// 	v1Router.Get("/error", handlerError)

// 	router.Mount("/v1", v1Router)

// 	srv := &http.Server{
// 		Handler: router,
// 		Addr:    fmt.Sprintf(":%s", portString),
// 	}
// 	log.Printf("Server listening on port %s", portString)
// 	err := srv.ListenAndServe()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// }
