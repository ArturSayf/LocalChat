// Writing a basic HTTP server is easy using the
// `net/http` package.
package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
)

type Message struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

var (
	messageID atomic.Int64
	messages  []Message
	mu        sync.Mutex
)

//go:embed index.html
var indexPage string

// A fundamental concept in `net/http` servers is
// *handlers*. A handler is an object implementing the
// `http.Handler` interface. A common way to write
// a handler is by using the `http.HandlerFunc` adapter
// on functions with the appropriate signature.

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, indexPage)
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Errorform", http.StatusBadRequest)
		return
	}
	id := int(messageID.Add(1))
	name := r.FormValue("username")
	text := r.FormValue("message")

	m := Message{
		ID:      id,
		Name:    name,
		Message: text,
	}

	mu.Lock()
	messages = append(messages, m)

	if len(messages) > 10 {
		messages = messages[1:]
	}
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mu.Lock()
	defer mu.Unlock()

	json.NewEncoder(w).Encode(messages)
}

func main() {

	// We register our handlers on server routes using the
	// `http.HandleFunc` convenience function. It sets up
	// the *default router* in the `net/http` package and
	// takes a function as an argument.
	http.HandleFunc("/", index)
	http.HandleFunc("/send-message", sendMessage)
	http.HandleFunc("/get-messages", getMessages)
	// Finally, we call the `ListenAndServe` with the port
	// and a handler. `nil` tells it to use the default
	// router we've just set up.
	http.ListenAndServe(":8090", nil)
}
