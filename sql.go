package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/go-sql-driver/mysql"
)

// Book struct represents a book entity
type Book struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Author string `json:"author"`
}

var db *sql.DB
var books []Book

func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:Sql@2024@tcp(127.0.0.1:3306)/dbname")
	if err != nil {
		log.Fatal(err)
	}

	// Create books table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS books (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			author VARCHAR(255) NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func getBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	books := fetchBooksFromDB()
	json.NewEncoder(w).Encode(books)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = deleteBookFromDB(id)
	if err != nil {
		http.Error(w, "Error deleting book", http.StatusInternalServerError)
		return
	}

	books := fetchBooksFromDB()
	json.NewEncoder(w).Encode(books)
}

func getBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	book, err := getBookFromDB(id)
	if err != nil {
		http.Error(w, "Error retrieving book", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(book)
}


func createBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	var newBook Book
	if err := json.NewDecoder(r.Body).Decode(&newBook); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	newBook.ID = strconv.Itoa(rand.Intn(1000000))
	err := insertBookToDB(newBook)
	if err != nil {
		http.Error(w, "Error creating book", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(newBook)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedBook Book
	if err := json.NewDecoder(r.Body).Decode(&updatedBook); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	updatedBook.ID = strconv.Itoa(id)
	err = updateBookInDB(updatedBook)
	if err != nil {
		http.Error(w, "Error updating book", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedBook)
}

func fetchBooksFromDB() []Book {
	rows, err := db.Query("SELECT id, name, author FROM books")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Name, &book.Author)
		if err != nil {
			log.Fatal(err)
		}
		books = append(books, book)
	}

	return books
}

func deleteBookFromDB(id int) error {
	_, err := db.Exec("DELETE FROM books WHERE id = ?", id)
	return err
}

func getBookFromDB(id int) (Book, error) {
	var book Book
	err := db.QueryRow("SELECT id, name, author FROM books WHERE id = ?", id).Scan(&book.ID, &book.Name, &book.Author)
	return book, err
}

func insertBookToDB(book Book) error {
	_, err := db.Exec("INSERT INTO books (name, author) VALUES (?, ?)", book.Name, book.Author)
	return err
}

func updateBookInDB(book Book) error {
	_, err := db.Exec("UPDATE books SET name = ?, author = ? WHERE id = ?", book.Name, book.Author, book.ID)
	return err
}

func main() {
	initDB()

	r := mux.NewRouter()

	books = append(books, Book{ID: "1", Name: "Geeta", Author: "Krishna"})
	books = append(books, Book{ID: "2", Name: "Life Journey", Author: "Steve Smith"})

	r.HandleFunc("/api/getbooks", getBooks).Methods("GET")
	r.HandleFunc("/api/createbooks", createBook).Methods("POST")
	r.HandleFunc("/api/getbooks/{id}", getBook).Methods("GET")
	r.HandleFunc("/api/updatebooks/{id}", updateBook).Methods("PUT")
	r.HandleFunc("/api/deletebooks/{id}", deleteBook).Methods("DELETE")

	fmt.Println("Starting server at port 8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}
