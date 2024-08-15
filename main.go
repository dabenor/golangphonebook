// Start server set up routes
package main

import (
	"golangphonebook/pkg/contacts"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	// C
	router.HandleFunc("/addContact", contacts.PutContact).Methods("PUT")
	// R
	router.HandleFunc("/getContacts", contacts.GetContacts).Methods("GET")
	// U
	router.HandleFunc("/updateContact", contacts.UpdateContact).Methods("POST")
	// D
	router.HandleFunc("/deleteContact/{id}", contacts.DeleteContact).Methods("DELETE")
	// Add router for dynamic routes
	http.Handle("/", router)

	http.ListenAndServe(":8080", nil)
}
