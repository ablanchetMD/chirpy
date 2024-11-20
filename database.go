package main

import (
	 "encoding/json"
	 "fmt"
	 "io"
	 "net/http"
	 "github.com/google/uuid"
	 "github.com/ablanchetMD/chirpy/internal/database"
	 "time"
	 "database/sql"
	// "sort"
	// "strconv"
	// "strings"
		
)


type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func mapChirpStruct(src database.Chirp) Chirp {
	return Chirp{
		ID:        src.ID,
		CreatedAt: src.CreatedAt,
		UpdatedAt: src.UpdatedAt,
		Body:      src.Body,
		UserID:    src.UserID,
	}
}

func handleCreateChirp(c *apiConfig, w http.ResponseWriter, r *http.Request) {
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
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	content, ok := requestData["body"]
	if !ok {		
		respondWithError(w, http.StatusBadRequest, "Missing body field")
		return
	}

	user_id, ok := requestData["user_id"]
	if !ok {		
		respondWithError(w, http.StatusBadRequest, "Missing user_id field")
		return
	}
	parsed_id, err := uuid.Parse(user_id)
	if err != nil {		
		respondWithError(w, http.StatusBadRequest, "Invalid user_id field")
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
	chirp, err := c.Db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body: cleaned_body,		
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: parsed_id,
	})
	if err != nil {
		
		fmt.Println("Error creating chirp: ", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp")
		return
	}
	
	// user.Password = nil
	respondWithJSON(w, http.StatusCreated, mapChirpStruct(chirp))
}

// func (db *DB) handleLogin(w http.ResponseWriter, r *http.Request) {
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "Error reading request body", http.StatusInternalServerError)
// 		respondWithError(w, http.StatusBadRequest, "No body in request")
// 		return
// 	}
// 	defer r.Body.Close()

// 	var requestData map[string]string
// 	err = json.Unmarshal(body, &requestData)
// 	if err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		respondWithError(w, http.StatusBadRequest, "Invalid request body")
// 		return
// 	}

// 	email, ok := requestData["email"]
// 	if !ok {
// 		respondWithError(w, http.StatusBadRequest, "Please include an email field")
// 		return
// 	}
// 	password, ok := requestData["password"]
// 	if !ok {
// 		respondWithError(w, http.StatusBadRequest, "Please include a password field")
// 		return
// 	}
// 	loggedUser, err := db.verifyLogin(email, password)

// 	if err != nil {
// 		respondWithError(w, http.StatusUnauthorized, "Password or email is invalid.")
// 		return
// 	}

// 	loggedUser.Password = nil

// 	respondWithJSON(w, http.StatusOK, loggedUser)

// }

func handleGetChirps(c *apiConfig, w http.ResponseWriter, r *http.Request) {
	

	chirps, err := c.Db.GetChirps(r.Context())
	if err != nil {
		fmt.Println("Error getting chirp: ", err)
		respondWithError(w, http.StatusInternalServerError, "Error getting chirps")
		return
	}
	var chirpStructs []Chirp
	for _, chirp := range chirps {
		chirpStructs = append(chirpStructs, mapChirpStruct(chirp))
	}
	respondWithJSON(w, http.StatusOK, chirpStructs)
}

func handleGetChirp(c *apiConfig, w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
    if len(id) == 0 {
        respondWithError(w, http.StatusBadRequest,"Chirp ID not provided")
        return
    }

	parsed_id, err := uuid.Parse(id)
	if err != nil {		
		respondWithError(w, http.StatusBadRequest, "Invalid id field")
		return
	}
    
	chirp, err := c.Db.GetChirp(r.Context(),parsed_id)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No chirp with that id: ", err)
			respondWithError(w, http.StatusNotFound, "No Chirp with that id")
			return
		} 
		fmt.Println("Error getting chirp: ", err)
		respondWithError(w, http.StatusInternalServerError, "Error getting chirp")
		return
	}
	 
	respondWithJSON(w, http.StatusOK, mapChirpStruct(chirp))
}

// func (db *DB) serverGetChirps(w http.ResponseWriter, r *http.Request) {
// 	pathParts := strings.Split(r.URL.Path, "/")
// 	if len(pathParts) < 4 || pathParts[3] == "" {
// 		// No ID provided, return all chirps
// 		chirps, err := db.GetChirps()
// 		if err != nil {
// 			respondWithError(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}
// 		respondWithJSON(w, http.StatusOK, chirps)
// 		return
// 	}
// 	// ID provided, try to parse and get the specific chirp
// 	id, err := strconv.Atoi(pathParts[3])
// 	if err != nil {
// 		respondWithError(w, http.StatusBadRequest, "No proper ID provided")
// 		return
// 	}

// 	chirps, err := db.GetChirp(id)
// 	if err != nil {
// 		respondWithError(w, http.StatusNotFound, "Chirp not found")
// 		return
// 	}

// 	respondWithJSON(w, http.StatusOK, chirps)

// }

// func (db *DB) GetChirps() ([]Chirp, error) {
// 	db.mux.RLock()
// 	defer db.mux.RUnlock()
// 	data, err := db.loadDB()
// 	if err != nil {
// 		return nil, err
// 	}
// 	// If no ID is provided, return all chirps

// 	var chirps []Chirp
// 	for _, chirp := range data.Chirps {
// 		chirps = append(chirps, chirp)
// 	}
// 	sort.Slice(chirps, func(i, j int) bool {
// 		return chirps[i].Id < chirps[j].Id
// 	})
// 	return chirps, nil

// }

// func (db *DB) GetChirp(id int) (Chirp, error) {
// 	db.mux.RLock()
// 	defer db.mux.RUnlock()
// 	chirp := Chirp{}
// 	data, err := db.loadDB()
// 	if err != nil {
// 		return chirp, err
// 	}

// 	chirp, ok := data.Chirps[id]
// 	if !ok {
// 		return chirp, fmt.Errorf("Chirp with ID %d not found", id)
// 	}
// 	return chirp, nil

// }
