package main

// Eldad Danladi
// Nathan Hislop

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Maou-Shimazu/2242-Final/internal/cookies"
)

var (
	ErrValueTooLong = errors.New("cookie value too long")
	ErrInvalidValue = errors.New("invalid cookie value")
)

var secret []byte

type User struct {
	Name string
	Age  int
}

func main() {
	// Decode the random 64-character hex string to give us a slice containing
	// 32 random bytes. For simplicity, I've hardcoded this hex string but in a
	// real application you should read it in at runtime from a command-line
	// flag or environment variable.
	gob.Register(&User{})

	var err error

	secret, err = hex.DecodeString("13d6b4dff8f84a10851021ec8608f814570d562c92fe6b5ec4c9f595bcb3234b")
	if err != nil {
		log.Fatal(err)
	}
	// start the server
	mux := http.NewServeMux()
	mux.HandleFunc("/set", setCookieHandler)
	mux.HandleFunc("/get", getCookieHandler)

	log.Print("Listening on :3000")
	err = http.ListenAndServe(":3000", mux)
	log.Fatal(err)
}

func setCookieHandler(w http.ResponseWriter, r *http.Request) {
	user := User{Name: "Maou", Age: 17}
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(&user)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "Cookie",
		Value:    buf.String(),
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	err = cookies.WriteEncrypted(w, cookie, secret)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("The cookie has been set!"))
}

func getCookieHandler(w http.ResponseWriter, r *http.Request) {
	gobEncodedValue, err := cookies.ReadEncrypted(r, "Cookie", secret)
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			http.Error(w, "a cookie was not found", http.StatusBadRequest)
		case errors.Is(err, ErrInvalidValue):
			http.Error(w, "the cookie value was invalid", http.StatusBadRequest)
		default:
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}
	var user User

	reader := strings.NewReader(gobEncodedValue)

	if err := gob.NewDecoder(reader).Decode(&user); err != nil {
		log.Println(err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Name: %q\n", user.Name)
	fmt.Fprintf(w, "Age: %d\n", user.Age)
}
