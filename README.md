# Phonebook API Documentation

Experimenting with Go to create a WebServer API with basic CRUD operations

## Setup

To build and test this project, run the following from the root directory
```bash
docker-compose up --build test
```
Next, close the docker build once the tests are complete by pressing Ctrl+C, and then removing the test containers with the following command

```bash
docker-compose down
```
To then run the resulting Docker build and expose port 8080, please run the following, also from the root directory

```bash
docker-compose up db phonebook
```

Then you can access the application by sending CURL requests to [http://localhost:8080/](http://localhost:8080/) from the terminal, via [Postman](https://www.postman.com/), or you can navigate to the same URL in your browser.


## Table of Contents

1. [Endpoints](#endpoints)
    - [Add Contact](#add-contact)
    - [Add Contacts](#add-contacts)
    - [Get Contacts](#get-contacts)
    - [Update Contact](#update-contact)
    - [Delete Contact](#delete-contact)
    - [Delete Contacts](#delete-contacts)

## Endpoints

---

### Add Contact

- **Endpoint**: `/addContact`
- **Method**: PUT
- **Description**: Creates multiple new contacts based on the provided JSON body.

#### Request Body

- An array of JSON objects representing the contacts to add. Each object should include at least the 'first_name' and 'phone' fields. Optional fields that can also be populated later using an [Update Contact](#update-contact) call are 'last_name' and 'address'

**Example Request Body**:

```json
{
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "address": "123 Main St"
}
```

**Responses:**
- 200 OK: Contact added successfully.
- 400 Bad Request: Invalid request body, first name and phone must be correctly defined.
- 500 Internal Server Error: Failed to insert contact into the database.


### Add Contacts

- **Endpoint**: `/addContacts`
- **Method**: PUT
- **Description**: Creates multiple new contacts based on the provided JSON array.

#### Request Body

- An array of JSON objects representing the contacts to add. Each object should include 'first_name', 'last_name', and 'phone' fields.

**Example Request Body**:

```json
[
  {
    "first_name": "John",
    "last_name": "Doe",
    "phone": "1234567890"
  },
  {
    "first_name": "Jane",
    "last_name": "Smith",
    "phone": "9876543210"
  }
]
```
