// Start server set up routes
package main

import (
	"golangphonebook/pkg/contacts"
	"net/http"
)

func main() {
	// C
	http.HandleFunc("/addContact", contacts.PutContact)
	// R
	http.HandleFunc("/getContacts", contacts.GetContacts)
	// U
	http.HandleFunc("/updateContact", contacts.UpdateContact)
	// D
	http.HandleFunc("/deleteContact", contacts.DeleteContact)

	http.ListenAndServe(":8080", nil)
}
