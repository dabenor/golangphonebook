// Start server set up routes
package main

import (
	"database/sql"
	"golangphonebook/internal"
	"golangphonebook/pkg/contacts"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	// Set up PostgreSQL connection
	connStr := "user=username password=password dbname=mydb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		internal.Logger.Error("DB connection init failed, shutting down: %s", err)
		return
	}
	defer db.Close()

	// Initialize the db interaction functions
	repo := contacts.NewSQLContactRepository(db)

	router := mux.NewRouter()
	// C
	router.HandleFunc("/addContact", func(w http.ResponseWriter, r *http.Request) { contacts.PutContact(w, r, repo) }).Methods("PUT")
	// R
	router.HandleFunc("/getContacts", func(w http.ResponseWriter, r *http.Request) { contacts.GetContacts(w, r, repo) }).Methods("GET")
	// U
	router.HandleFunc("/updateContact", func(w http.ResponseWriter, r *http.Request) { contacts.UpdateContact(w, r, repo) }).Methods("POST")
	// D
	router.HandleFunc("/deleteContact/{id}", func(w http.ResponseWriter, r *http.Request) { contacts.DeleteContact(w, r, repo) }).Methods("DELETE")
	// Add router for dynamic routes
	http.Handle("/", router)

	http.ListenAndServe(":8080", nil)
}
