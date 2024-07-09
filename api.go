package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	getEnv "github.com/joho/godotenv"
)

// function for generating JSON : 4
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// make handle func to return httpHandler : 4
type apiFunc func(w http.ResponseWriter, r *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

// 5
func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			// handle the error
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

// 1
type APIServer struct {
	listenAddr string
	store      Storage
}

// 2
func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

// server initialization: 6
func (s *APIServer) Run() {
	router := mux.NewRouter()
	// video 5 .1
	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin))
	// we need to wrap the handle bsc HandleFunc don't return
	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleGetAccountById), s.store))
	router.HandleFunc("/transfer", makeHTTPHandleFunc(s.handleTransfer))

	log.Println("JSON API run on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

// video 5.2
func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, req)
}

// handlers 3
func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}
	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

// GET /account
func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountById(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		idStr := mux.Vars(r)["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return fmt.Errorf("Invalid id %s", idStr)
		}
		log.Println(id)
		// get DB by id
		account, err := s.store.GetAccountByID(id)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, account)
		// use empty object for testing
		// return WriteJSON(w, http.StatusOK, &Account{})
	}
	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

// after writing createaccounrequest on types
func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountRequest := new(CreatedAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountRequest); err != nil {
		return err
	}

	// make account object
	account, err := NewAccount(createAccountRequest.FirstName, createAccountRequest.LastName, createAccountRequest.Password)
	// video 5.4
	if err != nil {
		return err
	}

	// put account on db
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	// video 4. 3
	tokenString, err := createJWT(account)
	if err != nil {
		return err
	}

	fmt.Println("JWT token: ", tokenString)

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}
	err = s.store.DeleteAccount(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := new(TransferRequest)
	err := json.NewDecoder(r.Body).Decode(transferReq)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return WriteJSON(w, http.StatusOK, transferReq)
}

func getId(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("Invalid id %s", idStr)
	}

	return id, nil
}

// video 4.3
// create jwt token when creating account
func createJWT(account *Account) (string, error) {
	// Create the Claims
	claims := &jwt.MapClaims{
		"ExpiresAt":     15000,
		"accountNumber": account.Number,
	}
	secret := os.Getenv("JWT_SECRET")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
	// fmt.Println(ss, err)
}

// video 4.5
func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
}

// video 4. JWT middleware
// video 4. 4 Storage to access db
func withJWTAuth(handlerfunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		fmt.Println("Using JWT middleware")

		tokenString := r.Header.Get("x-jwt-token")

		token, err := validateJWT(tokenString)
		if err != nil {
			permissionDenied(w)
			return
		}

		if !token.Valid {
			permissionDenied(w)
			return
		}
		// video 4. 5
		userID, err := getId(r)
		if err != nil {
			permissionDenied(w)
			return
		}

		account, err := s.GetAccountByID(userID)
		if err != nil {
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		// panic(reflect.TypeOf(claims["accountNumber"])) // this is float64
		if account.Number != int64(claims["accountNumber"].(float64)) { // conert the interface into float64
			permissionDenied(w)
			return
		}

		// fmt.Println(claims)

		handlerfunc(w, r)
	}
}

// video 4.2
func validateJWT(tokenString string) (*jwt.Token, error) {
	err := getEnv.Load()
	if err != nil {
		log.Fatal(err)
	}
	secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secret), nil
	})
}
