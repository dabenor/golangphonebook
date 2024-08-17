// Define structures and structure enforcement infrastructure
package contacts

import (
	"fmt"
	"golangphonebook/internal"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

type Contact struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement;index:idx_first_last,priority:3;index:idx_last_first,priority:3"` // Auto-incrementing primary key
	FirstName    string    `json:"first_name" validate:"required" gorm:"size:50;not null;index:idx_first_last,priority:1"`             // Index on FirstName with LastName and ID
	LastName     string    `json:"last_name" gorm:"size:50;index:idx_first_last,priority:2;index:idx_last_first,priority:1"`           // Index on LastName with FirstName and ID
	Phone        string    `json:"phone" validate:"required,customPhone" gorm:"size:20"`                                               // Phone field with validation and size constraint
	Address      string    `json:"address" gorm:"size:100;type:text"`                                                                  // Address field, stored as text in the database
	LastModified time.Time `json:"last_modified" gorm:"autoUpdateTime;index"`                                                          // Automatically updated on save

}

type ContactList struct {
	contacts []Contact
	count    int
}

func (c Contact) String() string {
	if c.LastModified.IsZero() {
		return fmt.Sprintf("Contact(ID=%d, FirstName=%s, LastName=%s, Phone=%s, Address=%s)",
			c.ID, c.FirstName, c.LastName, c.Phone, c.Address)
	} else {
		return fmt.Sprintf("Contact(ID=%d, FirstName=%s, LastName=%s, Phone=%s, Address=%s, LastModified=%s)",
			c.ID, c.FirstName, c.LastName, c.Phone, c.Address, c.LastModified)
	}
}

// DB interaction interface
type ContactRepository interface {
	AddContact(contact Contact) error
	GetContacts(page int) error
	GetAllContacts()
	UpdateContact(id int, contact Contact) error
	DeleteContact(id int) error
	GetContactCount() (int64, error)
}

// Structure validator
var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())

	validate.RegisterValidation("customPhone", regexValidator("^\\+?[0-9]{4,20}$"))

}

// return a validation function to check the field against the regex
func regexValidator(pattern string) validator.Func {
	return func(fl validator.FieldLevel) bool {
		// Compile the regular expression
		re, err := regexp.Compile(pattern)
		if err != nil {
			// Invalid regex pattern
			internal.Logger.Error("Invalid regex pattern on server side")
			return false
		}

		// Validate the field value against the regex pattern
		matches := re.MatchString(fl.Field().String())
		internal.Logger.Info(fmt.Sprintf("Validating field '%s' with value '%s': %v", fl.FieldName(), fl.Field().String(), matches))
		return matches
	}
}

// Addtl. structures as needed will go here
