package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ablanchetMD/chirpy/internal/auth"
	"github.com/google/uuid"

	"github.com/ablanchetMD/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func mapUserStruct(src database.User) User {
	return User{
		ID:        src.ID,
		CreatedAt: src.CreatedAt,
		UpdatedAt: src.UpdatedAt,
		Email:     src.Email,
	}
}

func handleCreateUser(c *apiConfig, w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
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

	email, ok := requestData["email"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Missing email field")
		return
	}
	password, ok := requestData["password"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Missing password field")
		return
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password")
		return
	}

	fmt.Println("Adding user to database: ", email)
	user, err := c.Db.CreateUser(r.Context(), database.CreateUserParams{
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Password:  hashedPassword,
	})
	if err != nil {
		fmt.Println("Error creating user: ", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating user")
		return
	}
	respondWithJSON(w, http.StatusCreated, mapUserStruct(user))
}

func handleReset(c *apiConfig, w http.ResponseWriter, r *http.Request) {
	if c.Platform != "dev" {
		respondWithError(w, http.StatusForbidden, "You are not authorized to use this function.")
		return
	}

	err := c.Db.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting users")
		return
	}
	respondWithJSON(w, http.StatusOK, "Users deleted")
}

func handleLogin(c *apiConfig, w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
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
	user, err := c.Db.GetUserByEmail(r.Context(), email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Password or email is invalid.")
		return
	}
	err = auth.CheckPasswordHash(password, user.Password)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Password or email is invalid.")
		return
	}

	// user.Password = nil

	respondWithJSON(w, http.StatusOK, mapUserStruct(user))

}

// func (db *DB) findEmail(email string) (User, error) {
// 	user := User{}
// 	data, err := db.loadDB()
// 	if err != nil {
// 		return user, err
// 	}
// 	for _, user := range data.Users {
// 		if user.Email == email {
// 			return user, nil
// 		}
// 	}
// 	return user, fmt.Errorf("User with email %s not found", email)
// }

// func (db *DB) verifyLogin(email, password string) (User, error) {
// 	db.mux.RLock()
// 	defer db.mux.RUnlock()
// 	user := User{}

// 	findUser, err := db.findEmail(email)

// 	if err != nil {
// 		return user, err
// 	}

// 	err = bcrypt.CompareHashAndPassword(findUser.Password, []byte(password))
// 	if err != nil {
// 		return user, fmt.Errorf("Password %s is incorrect for %s", password, email)
// 	}

// 	return findUser, nil

// }
