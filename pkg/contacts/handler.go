// handle HTTP function requests GET POST DELETE etc.
package contacts

import (
	"encoding/json"
	"fmt"
	"golangphonebook/internal"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func PutContact(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
	defer internal.Timer("PutContact")()

	contact, err := decodeBodyToContact(r)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Received invalid body in addContact method %s", err))
		http.Error(w, "Invalid request body, first name and phone must be correctly defined", http.StatusBadRequest)
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
	// State tracking for caching, since changes to DB we need to pull fresh data
	filterState.UpdateCache = true

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Contact added to DB successfully"))
	// Return an updated first page, maybe. Return contact itself. Or when user adds a contact, send them back to their origin page, or to page 1 of their contacts
}

func GetContacts(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
	pageStr := r.URL.Query().Get("page")
	sortByStr := r.URL.Query().Get("sort_by")

	var sortBy SortBy
	switch sortByStr {
	case "first_name":
		sortBy = SortByFirstName
	case "last_name":
		sortBy = SortByLastName
	case "last_modified":
		sortBy = SortByLastModified
	default:
		sortBy = SortByFirstName
	}

	filters := map[string]string{
		"first_name": r.URL.Query().Get("first_name"),
		"last_name":  r.URL.Query().Get("last_name"),
		"address":    r.URL.Query().Get("address"),
		"phone":      r.URL.Query().Get("phone"),
	}
	// For comparisons, check if changes to filter
	queryString := buildFilterQueryString(filters)

	// Get the filtered gorm query, total count of contacts that match that query
	query, totalCount, err := repo.FilterContacts(filters)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Failed to filter contacts: %v", err))
		http.Error(w, "Failed to filter contacts", http.StatusInternalServerError)
		return
	}

	totalPages := int((totalCount + 9) / 10)
	// Failsafe for out of bounds page numbers, some tolerance for invalid page number input (just default to 1)
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 || page > totalPages {
		page = 1
	}

	// Check if the filter or page has changed
	isInitialFetch := false
	if filterState.QueryString == queryString && page == filterState.CachedPage && !filterState.UpdateCache {
		// If the filter is the same and page is the same, serve from cache
		if len(filterState.Cache) > 0 {
			cachedContacts := filterState.Cache[:10]
			filterState.Cache = filterState.Cache[10:]

			paginatedContacts := PaginatedContacts{
				Contacts:    cachedContacts,
				TotalPages:  filterState.TotalPages,
				CurrentPage: page,
				TotalCount:  filterState.TotalCount,
			}

			response, err := json.Marshal(paginatedContacts)
			if err != nil {
				internal.Logger.Error(fmt.Sprintf("Failed to serialize contacts: %v", err))
				http.Error(w, "Failed to serialize contacts", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)

			// Start goroutine to prefetch the next set of contacts
			go func() {
				contacts, err := repo.SearchContacts(filterState.Query, page+1, sortBy, false)
				if err == nil && len(contacts) > 0 {
					internal.Logger.Info("Cache updated successfully")
				} else {
					internal.Logger.Error("Failed to update cache, setting cache to try updating again with next call")
					filterState.UpdateCache = true
				}
			}()

			return
		}

	} else { // If it's a new fetch continue below
		isInitialFetch = true
		filterState.Query = query
		filterState.QueryString = queryString
		filterState.CachedPage = page + 1
		filterState.UpdateCache = false
	}

	// Get the contacts for the specified page using SearchContacts
	contacts, err := repo.SearchContacts(query, page, sortBy, isInitialFetch)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Failed to search contacts: %v", err))
		http.Error(w, "Failed to search contacts", http.StatusInternalServerError)
		return
	}

	// Store pagination info in filterState
	filterState.TotalPages = totalPages
	filterState.TotalCount = totalCount

	// Cache the next set of contacts if it's the initial fetch
	if isInitialFetch && len(contacts) > 10 {
		filterState.Cache = contacts[10:]
		contacts = contacts[:10]
	}

	// Construct the PaginatedContacts object
	paginatedContacts := PaginatedContacts{
		Contacts:    contacts,
		TotalPages:  totalPages,
		CurrentPage: page,
		TotalCount:  totalCount,
	}

	// Serialize the PaginatedContacts object to JSON
	response, err := json.Marshal(paginatedContacts)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Failed to serialize contacts: %v", err))
		http.Error(w, "Failed to serialize contacts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

	// Start goroutine to prefetch the next set of contacts
	if !isInitialFetch {
		go func() {
			contacts, err := repo.SearchContacts(query, page+1, sortBy, false)
			if err == nil && len(contacts) > 0 {
				filterState.Cache = append(filterState.Cache, contacts...)
			}
		}()
	}
}

func GetAllContacts(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
	repo.GetAllContacts()
}

func UpdateContact(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
	// Extract ID from  URL path /deleteContact/{id}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(w, "Invalid ID, IDs can only be integers", http.StatusBadRequest)
		return
	}
	internal.Logger.Info(fmt.Sprintf("ID to update detected as %d", id))

	contact, err := decodeBodyToContact(r)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Received invalid body in updateContact method %s", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	internal.Logger.Info(fmt.Sprintf("Received valid body in updateContact method %s", contact))

	// Update contact in db
	err = repo.UpdateContact(id, *contact)
	if err != nil {
		// Handle duplicate data error
		if err.Error() == "another contact with the same first name, last name, and phone number already exists" {
			internal.Logger.Error(fmt.Sprintf("Duplicate data error: %s", err))
			http.Error(w, "Duplicate contact with the same first name, last name, and phone number already exists", http.StatusBadRequest)
			return
		}
		// Handle other errors as internal server errors
		internal.Logger.Error(fmt.Sprintf("Failed to update contact to db: %s", err))
		http.Error(w, "Failed to update contact due to an internal server error", http.StatusInternalServerError)
		return
	}

	filterState.UpdateCache = true

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Contact updated successfully"))
}

func DeleteContact(w http.ResponseWriter, r *http.Request, repo ContactRepository) {
	defer internal.Timer("DeleteContact")()

	// Extract ID from  URL path /deleteContact/{id}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(w, "Invalid ID, IDs can only be integers", http.StatusBadRequest)
		return
	}
	internal.Logger.Info(fmt.Sprintf("ID to delete detected as %d", id))

	err = repo.DeleteContact(id)
	if err != nil {
		if err.Error() == "no contact found with the given ID" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete contact", http.StatusInternalServerError)
		}
		return
	}
	filterState.UpdateCache = true

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
		return nil, fmt.Errorf("Err(s):\n%+v", err)
	}
	// Return Contact object
	return &contact, nil
}

func buildFilterQueryString(filters map[string]string) string {
	var queryParts []string

	// Iterate over filters
	for key, value := range filters {
		if value != "" {
			// Build the query part for the current filter
			queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Sort query parts for consistent order
	sort.Strings(queryParts)

	// Join all parts into a string
	return strings.Join(queryParts, "&")
}
