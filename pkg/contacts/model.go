// Define structures
package contacts

import (
	"github.com/go-playground/validator/v10"
)

type Contact struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone" validate:"required,regexp=^\\+?[0-9]{4,20}$"`
	Address   string `json:"address"`
}

type ContactList struct {
	contacts []Contact
	count    int
}

// Structure validator
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Addtl. structures as needed will go here
