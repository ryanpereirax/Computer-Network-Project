package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
)

type Movie struct {
	gorm.Model
	Title       string
	Description string
	Year        int
}

type Review struct {
	gorm.Model
	MovieID uint
	UserID  uint
	Rating  int
	Comment string
}

type User struct {
	gorm.Model
	Name     string
	Email    string `gorm:"unique"`
	Password string
}

func main() {
	router := mux.NewRouter()

	db, err := gorm.Open("postgres", "host=localhost port=5432 user=postgres dbname=movie_review_api password=postgres sslmode=disable")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	db.AutoMigrate(&Movie{}, &Review{}, &User{})

	router.HandleFunc("/movies", createMovieHandler(db)).Methods("POST")
	router.HandleFunc("/movies/{id}", getMovieHandler(db)).Methods("GET")
	router.HandleFunc("/movies", getMoviesHandler(db)).Methods("GET")
	router.HandleFunc("/reviews", createReviewHandler(db)).Methods("POST")
	router.HandleFunc("/reviews/{id}", getReviewHandler(db)).Methods("GET")
	router.HandleFunc("/reviews", getReviewsHandler(db)).Methods("GET")
	router.HandleFunc("/reviews/{id}", updateReviewHandler(db)).Methods("PUT")
	router.HandleFunc("/reviews/{id}", deleteReviewHandler(db)).Methods("DELETE")
	router.HandleFunc("/users", createUserHandler(db)).Methods("POST")
	router.HandleFunc("/users/{id}", getUserHandler(db)).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func createMovieHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var movie Movie
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&movie); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer r.Body.Close()

		db.Create(&movie)
		respondWithJSON(w, http.StatusCreated, movie)
	}
}

func getMovieHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid movie ID")
			return
		}

		var movie Movie
		if db.First(&movie, id).RecordNotFound() {
			respondWithError(w, http.StatusNotFound, "Movie not found")
