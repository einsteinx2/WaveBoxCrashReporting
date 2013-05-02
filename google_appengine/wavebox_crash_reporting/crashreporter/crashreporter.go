package redirector

import (
	"fmt"
	"net/http"
	"appengine"
	"appengine/datastore"
	"time"
)

/*
 * Errors
 */

type ErrorCode int32
const (
	ErrorCode_DatabaseError 	= iota
	ErrorCode_MissingParameter
)

func errorMessage(c ErrorCode) (string) {
	switch c {
	case ErrorCode_DatabaseError:
		return "Database Error"
	case ErrorCode_MissingParameter:
		return "Missing Parameter"
	}
	return ""
}

/*
 * Datastore
 */

type CrashRecord struct {
	Timestamp 	int64
	Exception   string
}

// Put a new host url record into datastore
func storeCrashRecord(record *CrashRecord, r *http.Request) (error) {

	c := appengine.NewContext(r)
	k := datastore.NewIncompleteKey(c, "crash", nil)

	_, err := datastore.Put(c, k, record)
	return err
}

/*
 * Header writing
 */

// Send the not implemented header, used for request types other than GET and POST
func unimplemented(w http.ResponseWriter, r *http.Request) {

	header := w.Header()
	header.Set("Allow", "GET, POST")
	header.Set("Content-type", "text/html")
	w.WriteHeader(501)
}

// Send the registration success message
func success(w http.ResponseWriter, r *http.Request) {

	header := w.Header()
	header.Set("Content-type", "application/json")
	fmt.Fprint(w, "{\"success\":true}")
}

// Send the registration failure message
func failure(w http.ResponseWriter, r *http.Request, c ErrorCode) {

	header := w.Header()
    header.Set("Content-type", "application/json")
    fmt.Fprintf(w, "{\"success\":false,\"errorCode\":%d,\"errorMessage\":\"%s\"}", c, errorMessage(c)) 
}

/*
 * Request Handling
 */

func init() {

	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {

	// Make sure this is a GET or POST request
	if r.Method == "GET" || r.Method == "POST" {
		saveCrashReport(w, r);
	} else {
		unimplemented(w, r)
	}
}

func saveCrashReport(w http.ResponseWriter, r *http.Request) {

	// Parse the query parameters
	r.ParseForm()
	exception := r.Form.Get("exception")

	// Make sure all parameter exists
	if len(exception) == 0 {

		failure(w, r, ErrorCode_MissingParameter)

	// Record the crash record
	} else {

		// Create a new record for this crash
		record := CrashRecord {
			Timestamp: time.Now().Unix(),
			Exception: exception,
		}

		// Store the crash
		if err := storeCrashRecord(&record, r); err != nil {
			failure(w, r, ErrorCode_DatabaseError)
		} else {
			success(w, r)
		}
	} 
}