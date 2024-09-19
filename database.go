package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func NewDB(path string) (*DB, error) {
	current_directory, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return &DB{path: current_directory + "/" + path, mux: &sync.RWMutex{}}, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	data, err := db.loadDB()
	chirp := Chirp{}
	if err != nil {
		// If the database file is empty, we initialize the Chirps map
		if err.Error() == "database file is empty" {
			data.Chirps = make(map[int]Chirp)
		} else {
			return chirp, err
		}
	}
	// Find the next ID
	nextID := 1
	for id := range data.Chirps {
		if id >= nextID {
			nextID = id + 1
		}
	}
	chirp = Chirp{Id: nextID, Body: body}
	data.Chirps[nextID] = chirp
	err = db.writeDB(data)
	if err != nil {
		return chirp, err
	}
	return chirp, nil

}

func (db *DB) handleChirps(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		db.serverGetChirps(w, r)
	case http.MethodPost:
		db.serverPostChirps(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (db *DB) handleLogin(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		respondWithError(w, http.StatusBadRequest, "No body in request")
		return
	}
	defer r.Body.Close()

	var requestData map[string]string
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	email, ok := requestData["email"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Please include an email field")
		return
	}
	password, ok := requestData["password"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Please include a password field")
		return
	}
	loggedUser, err := db.verifyLogin(email, password)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Password or email is invalid.")
		return
	}

	loggedUser.Password = nil

	respondWithJSON(w, http.StatusOK, loggedUser)

}

func (db *DB) serverPostChirps(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		respondWithError(w, http.StatusBadRequest, "No body in request")
		return
	}
	defer r.Body.Close()

	var requestData map[string]string
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	//convert body to Go object
	content, ok := requestData["body"]
	if !ok {
		http.Error(w, "Missing content field", http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Missing content field")
		return
	}
	replacementWords := []string{"kerfuffle", "sharbert", "fornax"}
	cleaned_body := replaceWords(string(content), replacementWords, "****")
	bodyLength := len(content)

	if bodyLength > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	fmt.Println("Adding chirp to database: ", cleaned_body)
	chirp, err := db.CreateChirp(cleaned_body)
	if err != nil {
		fmt.Println("Error creating chirp: ", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp")
		return
	}
	respondWithJSON(w, http.StatusCreated, chirp)

}



func (db *DB) serverGetChirps(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 || pathParts[3] == "" {
		// No ID provided, return all chirps
		chirps, err := db.GetChirps()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, chirps)
		return
	}
	// ID provided, try to parse and get the specific chirp
	id, err := strconv.Atoi(pathParts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "No proper ID provided")
		return
	}

	chirps, err := db.GetChirp(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}

	respondWithJSON(w, http.StatusOK, chirps)

}

func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	data, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	// If no ID is provided, return all chirps

	var chirps []Chirp
	for _, chirp := range data.Chirps {
		chirps = append(chirps, chirp)
	}
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].Id < chirps[j].Id
	})
	return chirps, nil

}

func (db *DB) GetChirp(id int) (Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	chirp := Chirp{}
	data, err := db.loadDB()
	if err != nil {
		return chirp, err
	}

	chirp, ok := data.Chirps[id]
	if !ok {
		return chirp, fmt.Errorf("Chirp with ID %d not found", id)
	}
	return chirp, nil

}

func (db *DB) ensureDB() error {
	if _, err := os.Stat(db.path); os.IsNotExist(err) {
		fmt.Println("Creating new database at: ", db.path)
		os.Create(db.path)
		if _, err := os.Stat(db.path); os.IsNotExist(err) {
			//Something went wrong?
			return err
		}
	}
	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	// db.mux.RLock()
	// defer db.mux.RUnlock()
	ensureDBErr := db.ensureDB()
	if ensureDBErr != nil {
		return DBStructure{}, ensureDBErr
	}
	database := DBStructure{}
	file, err := os.Open(db.path)
	if err != nil {
		//There was an error in the database, so we return an empty one.
		fmt.Println("Error opening database: ", err)
		return database, err
	}
	defer file.Close()

	// Check if the file is empty
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return database, err
	}

	if fileInfo.Size() == 0 {
		fmt.Println("Database file is empty")
		//If it's empty, we should still return it but not try to decode it.
		return database, fmt.Errorf("database file is empty")
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&database)
	if err != nil {
		//There was an error in the database, so we return an empty one.
		return database, err
	}

	return database, nil
}

func (db *DB) writeDB(dbStruct DBStructure) error {
	// db.mux.Lock()
	// defer db.mux.Unlock()
	file, err := os.Create(db.path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(dbStruct)
	if err != nil {
		return err
	}

	return nil
}
