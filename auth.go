package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password []byte `json:"password,omitempty"`
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	data, err := db.loadDB()
	user := User{}
	if err != nil {
		// If the database file is empty, we initialize the Chirps map
		if err.Error() == "database file is empty" {
			data.Users = make(map[int]User)
		} else {
			return user, err
		}
	}
	// Find the next ID
	nextID := 1
	for id := range data.Users {
		if id >= nextID {
			nextID = id + 1
		}
	}
	if len(data.Users) == 0 {
		data.Users = make(map[int]User)
	}

	findEmail, err := db.findEmail(email)
	if findEmail.Email != "" {
		return user, fmt.Errorf("User with email %s already exists", email)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return user, err
	}

	user = User{Id: nextID, Email: email, Password: hashedPassword}
	data.Users[nextID] = user
	err = db.writeDB(data)
	if err != nil {
		return user, err
	}
	return user, nil

}

func (db *DB) findEmail(email string) (User, error) {
	user := User{}
	data, err := db.loadDB()
	if err != nil {
		return user, err
	}
	for _, user := range data.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return user, fmt.Errorf("User with email %s not found", email)
}

func (db *DB) verifyLogin(email, password string) (User, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	user := User{}

	findUser, err := db.findEmail(email)

	if err != nil {
		return user, err
	}

	err = bcrypt.CompareHashAndPassword(findUser.Password, []byte(password))
	if err != nil {
		return user, fmt.Errorf("Password %s is incorrect for %s", password, email)
	}

	return findUser, nil

}

func (db *DB) handleUsers(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Missing email field", http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Missing email field")
		return
	}
	password, ok := requestData["password"]
	if !ok {
		http.Error(w, "Missing password field", http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Missing password field")
		return
	}
	fmt.Println("Adding user to database: ", email)
	user, err := db.CreateUser(email, password)
	if err != nil {
		fmt.Println("Error creating user: ", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating user")
		return
	}
	user.Password = nil
	respondWithJSON(w, http.StatusCreated, user)
}
