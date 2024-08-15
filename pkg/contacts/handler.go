// handle HTTP function requests GET POST DELETE etc.
package contacts

import (
	"encoding/json"
	"fmt"
	"golangphonebook/internal"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator"
)

func PutContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contact, err := decodeBodyToContact(r)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Received invalid body in addContact method %s", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	} else {
		internal.Logger.Info(fmt.Sprintf("Received valid body in addContact method %s", contact))
	}

	err = addContact(*contact)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert contact to db with error %s", err), http.StatusInternalServerError)
	}

	// Return an updated first page, maybe. Return contact itself. Or when user adds a contact, send them back to their origin page, or to page 1 of their contacts
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

	contact, err := decodeBodyToContact(r)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Received invalid body in updateContact method %s", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	} else {
		// Require contact ID to address potential duplicate entries in the backend, need to know which contact we're updating
		if contact.ID == 0 {
			internal.Logger.Error("Missing Contact ID in request to update contact")
			http.Error(w, "Invalid request body, missing ID for contact update", http.StatusBadRequest)
			return
		}
		internal.Logger.Info(fmt.Sprintf("Received valid body in updateContact method %s", contact))
	}

	err = updateContact(*contact)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update contact to db with error %s", err), http.StatusInternalServerError)
	}
}

func DeleteContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from  URL path /deleteContact/{id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL, ID is missing", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[2])

	if err != nil {
		http.Error(w, "Invalid ID, IDs can only be integers", http.StatusBadRequest)
		return
	}
	internal.Logger.Info(fmt.Sprintf("ID detected is %d", id))

	deleteContact(id)
}

// Decode JSON body into a Contact
func decodeBodyToContact(r *http.Request) (*Contact, error) {
	// Read body from request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Unable to read request body: %v", err))
		return nil, fmt.Errorf("unable to read request body: %v", err)
	}
	defer r.Body.Close()
	internal.Logger.Info(fmt.Sprintf("Received JSON: %s", string(body)))

	// Decode JSON body into a Contact
	var contact Contact
	if err := json.Unmarshal(body, &contact); err != nil {
		internal.Logger.Error(fmt.Sprintf("Unable to unmarshal JSON into Contact: %v", err))
		return nil, fmt.Errorf("unable to unmarshal JSON into Contact: %v", err)
	}

	// Validate the Contact struct
	err = validate.Struct(contact)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				// Log and return specific validation error
				internal.Logger.Warn(fmt.Sprintf("Validation failed for field '%s': %v", fieldError.Field(), fieldError.Tag()))
				return nil, fmt.Errorf("validation failed for field '%s': %v", fieldError.Field(), fieldError.Tag())
			}
		}
	}
	// Return Contact object
	return &contact, nil
}
