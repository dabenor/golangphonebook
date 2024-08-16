// handle HTTP function requests GET POST DELETE etc.
package contacts

import (
	"encoding/json"
	"fmt"
	"golangphonebook/internal"
	"io"
	"net/http"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
)

func PutContact(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
	contact, err := decodeBodyToContact(r)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Received invalid body in addContact method %s", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	} else {
		internal.Logger.Info(fmt.Sprintf("Received valid body in addContact method %s", contact))
	}

	err = repo.AddContact(*contact)
	if err != nil {
		if err.Error() == "contact with the same full name and phone number already exists" {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, fmt.Sprintf("Failed to insert contact to db with error %v", err), http.StatusInternalServerError)
		}
		return
	}

	internal.Logger.Info("Contact added to DB successfully")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Contact added to DB successfully"))

	// Return an updated first page, maybe. Return contact itself. Or when user adds a contact, send them back to their origin page, or to page 1 of their contacts
}

func GetContacts(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
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

	repo.GetContacts(page)
}

func GetAllContacts(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
	repo.GetAllContacts()
}

func UpdateContact(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
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

	err = repo.UpdateContact(*contact)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update contact to db with error %s", err), http.StatusInternalServerError)
	}
}

func DeleteContact(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
	// Extract ID from  URL path /deleteContact/{id}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(w, "Invalid ID, IDs can only be integers", http.StatusBadRequest)
		return
	}
	internal.Logger.Info(fmt.Sprintf("ID detected is %d", id))

	err = repo.DeleteContact(id)
	if err != nil {
		if err.Error() == "no contact found with the given ID" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete contact", http.StatusInternalServerError)
		}
		return
	}

	internal.Logger.Info("Contact deleted successfully")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Contact deleted successfully"))
}

// Helper method(s)
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
