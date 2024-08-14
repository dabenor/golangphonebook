// handle HTTP function requests GET POST DELETE etc.
package contacts

import (
	"encoding/json"
	"golangphonebook/internal"
	"net/http"
	"strconv"
)

func PutContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil && err.Error() != "EOF" {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
}

func GetContacts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// default page is page 1
	page := 1

	pageHeader := r.Header.Get("page")
	if parsedPage, err := strconv.Atoi(pageHeader); err == nil {
		if parsedPage >= 1 && parsedPage <= getSize()/11+1 {
			page = parsedPage
		} else {
			internal.Logger.Warn("Invalid page number, defaulting to page 1")
		}
	}

	getContacts(page)
}

func UpdateContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func DeleteContacts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
