// Define structures
package contacts

type Contact struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
}

type ContactList struct {
	contacts []Contact
	count    int
}

// Addtl. structures as needed will go here
