// Handle CRUD ops for the database
package contacts

import (
	"errors"
	"fmt"
	"golangphonebook/internal"

	"gorm.io/gorm"
)

var filterState FilterState

type SQLContactRepository struct {
	DB *gorm.DB
}

// NewSQLContactRepository creates a new instance of SQLContactRepository
func NewSQLContactRepository(db *gorm.DB) *SQLContactRepository {
	return &SQLContactRepository{DB: db}
}

func (repo *SQLContactRepository) AddContact(contact Contact) error {
	var existingContact Contact

	// Check if a contact with the same FirstName, LastName and Phone already exists
	err := repo.DB.Where("first_name = ? AND last_name = ? AND phone = ?", contact.FirstName, contact.LastName, contact.Phone).First(&existingContact).Error
	if err == nil {
		// Contact already exists
		internal.Logger.Warn("contact with the same full name and phone number already exists")
		return errors.New("contact with the same full name and phone number already exists")
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Contact does not exist
		err = repo.DB.Create(&contact).Error
		return err
	} else {
		// We got some other error
		return err

	}
}

func (repo *SQLContactRepository) GetContacts(page int) error {
	internal.Logger.Info("Made it to getContacts")
	return nil

}

func (repo *SQLContactRepository) FilterContacts(filters map[string]string) (*gorm.DB, int64, error) {
	// Build the query based on filters
	query := repo.DB.Model(&Contact{})

	if firstName, exists := filters["first_name"]; exists {
		query = query.Where("first_name LIKE ?", "%"+firstName+"%")
	}

	if lastName, exists := filters["last_name"]; exists {
		query = query.Where("last_name LIKE ?", "%"+lastName+"%")
	}

	if address, exists := filters["address"]; exists {
		query = query.Where("address LIKE ?", "%"+address+"%")
	}

	if phone, exists := filters["phone"]; exists {
		query = query.Where("phone LIKE ?", "%"+phone+"%")
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	return query, count, nil
}

func (repo *SQLContactRepository) SearchContacts(query *gorm.DB, page int, sortBy SortBy, initialFetch bool) ([]Contact, error) {
	var contacts []Contact
	limit := 10
	if initialFetch {
		limit = 20 // Fetch 20 records initially
	}
	offset := (page - 1) * 10

	// Determine the sort order
	switch sortBy {
	case SortByFirstName:
		query = query.Order("first_name ASC")
	case SortByLastName:
		query = query.Order("last_name ASC")
	case SortByLastModified:
		query = query.Order("last_modified DESC")
	default:
		query = query.Order("first_name ASC") // Default sorting
	}

	// Retrieve the contacts with pagination
	err := query.Limit(limit).Offset(offset).Find(&contacts).Error
	if err != nil {
		return nil, err
	}

	// Update filter state
	if initialFetch {
		if len(contacts) > 10 {
			// Don't cache any of these if they don't exist
			filterState.Cache = contacts[10:] // Cache the next 10 records
			filterState.CachedPage = page + 1 // We cached the next page
		}
		contacts = contacts[:10] // Serve the first 10 records
	}

	return contacts, nil
}

// TODO: remove this method
func (repo *SQLContactRepository) GetAllContacts() {
	var contacts []Contact

	// Query the database to retrieve all contacts
	err := repo.DB.Find(&contacts).Error
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Error on getting contacts %v", err))
	}

	internal.Logger.Info(fmt.Sprintf("Retrieved %d contacts", len(contacts)))
	for _, contact := range contacts {
		fmt.Println(contact)
	}
}

func (repo *SQLContactRepository) UpdateContact(id int, updatedContact Contact) error {
	// Check if contact exists
	var existingContact Contact
	err := repo.DB.First(&existingContact, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("contact not found")
		}
		return err
	}

	// Check for duplicate contact
	var duplicateContact Contact
	err = repo.DB.Where("first_name = ? AND last_name = ? AND id != ? AND phone = ?",
		updatedContact.FirstName, updatedContact.LastName, id, updatedContact.Phone).First(&duplicateContact).Error
	if err == nil {
		// Duplicate exists
		return errors.New("another contact with the same first name, last name, and phone number already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Some other error
		return err
	}
	// Update fields
	if updatedContact.FirstName != "" {
		existingContact.FirstName = updatedContact.FirstName
	}
	if updatedContact.LastName != "" {
		existingContact.LastName = updatedContact.LastName
	}
	if updatedContact.Phone != "" {
		existingContact.Phone = updatedContact.Phone
	}
	if updatedContact.Address != "" {
		existingContact.Address = updatedContact.Address
	}

	// Save contact back to db
	err = repo.DB.Save(&existingContact).Error
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Encountered err while saving updated contact back to DB: %v", err))
		return err
	}

	internal.Logger.Info(fmt.Sprintf("Contact with ID %d updated successfully", id))
	return nil
}

func (repo *SQLContactRepository) DeleteContact(id int) error {
	result := repo.DB.Delete(&Contact{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		internal.Logger.Error(fmt.Sprintf("no contact found with ID: %d", id))
		return errors.New("no contact found with the given ID")
	}

	internal.Logger.Info(fmt.Sprintf("Contact deleted successfully, %d row(s) affected", result.RowsAffected))

	return nil
}

// Helper methods
func (repo *SQLContactRepository) GetContactCount() (int64, error) {
	var count int64
	err := repo.DB.Model(&Contact{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
