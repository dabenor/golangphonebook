// Start server set up routes
package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"golangphonebook/db"
	"golangphonebook/internal"
	"golangphonebook/pkg/contacts"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {

	db, err := db.DBInit()
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("DB connection init failed, shutting down: %s", err))
		return
	}

	// Initialize the db interaction functions
	repo := contacts.NewSQLContactRepository(db)

	router := mux.NewRouter()
	// C
	router.HandleFunc("/addContact", func(w http.ResponseWriter, r *http.Request) { contacts.PutContact(w, r, repo) }).Methods("PUT")
	router.HandleFunc("/addContacts", func(w http.ResponseWriter, r *http.Request) { contacts.PutContacts(w, r, repo) }).Methods("PUT")
	// R
	router.HandleFunc("/getContacts", func(w http.ResponseWriter, r *http.Request) { contacts.GetContacts(w, r, repo) }).Methods("GET")
	// U
	router.HandleFunc("/updateContact/{id}", func(w http.ResponseWriter, r *http.Request) { contacts.UpdateContact(w, r, repo) }).Methods("POST")
	// D
	router.HandleFunc("/deleteContact/{id}", func(w http.ResponseWriter, r *http.Request) { contacts.DeleteContact(w, r, repo) }).Methods("DELETE")
	router.HandleFunc("/deleteContacts", func(w http.ResponseWriter, r *http.Request) { contacts.DeleteContacts(w, r, repo) }).Methods("DELETE")
	// // Add router for dynamic routes
	// http.Handle("/", router)

	// Load the server's certificate and private key
	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		log.Fatalf("Failed to load server certificate and key: %v", err)
	}

	clientCACert, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatalf("Failed to load client CA certificate: %v", err)
	}

	clientCAPool := x509.NewCertPool()
	clientCAPool.AppendCertsFromPEM(clientCACert)

	// Configure TLS with client certificate verification
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    clientCAPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	server := &http.Server{
		Addr:      ":8443",
		Handler:   router,
		TLSConfig: tlsConfig,
	}

	internal.Logger.Info("Ready to take secure requests on https://localhost:8443\n")
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatalf("Failed to start HTTPS server: %v", err)
	}

}
