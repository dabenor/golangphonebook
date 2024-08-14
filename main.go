// Start server set up routes
package main

import (
	"golangphonebook/pkg/contacts"
	"net/http"
)

func main() {
	// C
	http.HandleFunc("/", contacts.PutContact)
	// R
	http.HandleFunc("/contacts", contacts.GetContacts)
	// U

	// D
	http.HandleFunc("/delete", contacts.DeleteContacts)
	http.ListenAndServe(":8080", nil)
}
