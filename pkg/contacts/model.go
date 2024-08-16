// Define structures and structure enforcement infrastructure
package contacts

import (
	"fmt"
	"golangphonebook/internal"
	"regexp"

	"github.com/go-playground/validator/v10"
)

type Contact struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone" validate:"required,customPhone"`
	Address   string `json:"address"`
}

type ContactList struct {
	contacts []Contact
	count    int
}

func (c Contact) String() string {
	return fmt.Sprintf("Contact(ID=%d, FirstName=%s, LastName=%s, Phone=%s, Address=%s)",
		c.ID, c.FirstName, c.LastName, c.Phone, c.Address)
}

// Service interface
type ContactDBService struct {
	repo ContactRepository
}

// DB interaction interface
type ContactRepository interface {
	AddContact(contact Contact) error
	GetContacts(page int) error
	UpdateContact(contact Contact) error
	DeleteContact(id int) error
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
		return re.MatchString(fl.Field().String())
	}
}

// Addtl. structures as needed will go here
