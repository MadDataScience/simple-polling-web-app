package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Page struct {
	Title string
	Body  []byte
}

type User struct {
	Email           string
	Password        string
	ConfirmPassword string
}

type ErrorResponse struct {
    Err string
}

func (u *User) save() error {
	if u.Password != u.ConfirmPassword {
		return errors.New("passwords do not match, try again.")
	}
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		return err
	}

	emailBytes := []byte(u.Email)
	passwordBytes := []byte(u.Password)
	emailMD5 := md5.Sum(emailBytes)
	passwordMD5 := md5.Sum(passwordBytes)
	hashedEmail := hex.EncodeToString(emailMD5[:])
	hashedPassword := hex.EncodeToString(passwordMD5[:])

	fmt.Println("hashedEmail:", hashedEmail)
	fmt.Println("hashedPassword:", hashedPassword)

	statement, err :=
		db.Prepare("CREATE TABLE IF NOT EXISTS users (email VARCHAR(320) PRIMARY KEY, hashedEmail CHAR(32), hashedPassword CHAR(32))")
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	statement, err = db.Prepare("INSERT INTO users (email, hashedEmail, hashedPassword) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = statement.Exec(u.Email, hashedEmail, hashedPassword)
	if err != nil {
		return err
	}
	rows, err := db.Query("SELECT email, hashedEmail, hashedPassword FROM users")
	if err != nil {
		return err
	}
	var email string
	for rows.Next() {
		rows.Scan(&email, &hashedEmail, &hashedPassword)
		fmt.Println(email + ": " + hashedEmail + " " + hashedPassword)
	}
	return nil
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
    user := &User{}
    json.NewDecoder(r.Body).Decode(user)

    pass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        fmt.Println(err)
        err := ErrorResponse{
            Err: "Password Encryption failed",
        }
        json.NewEncoder(w).Encode(err)
    }

    user.Password = string(pass)

    createdUser := db.Create(user)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, time.Now().Format("2006-01-02 15:04:05"))
}

func newUserHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("new-user.html")
	t.Execute(w, nil)
}

// func (u *User) login() (string, error) {

// }

func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.gtpl")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		// logic part of log in
		fmt.Println("email:", r.Form["email"])
		fmt.Println("password:", r.Form["password"])
	}
}

func saveUserHandler(w http.ResponseWriter, r *http.Request) {
	u := &User{
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirm-password"),
	}
	err := u.save()
	if err != nil {
		fmt.Print(err)
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func main() {
	http.HandleFunc("/", MainHandler)
	http.HandleFunc("/new-user", newUserHandler)
	http.HandleFunc("/save-user", saveUserHandler)
	// http.HandleFunc("/login", login)

	fmt.Println("Listening on port 5050...")

	http.ListenAndServe(":5050", nil)
}
