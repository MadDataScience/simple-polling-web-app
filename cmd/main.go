package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"time"

	// "encoding/json"
	"errors"
	"fmt"
	"html/template"

	// "io"
	"io/ioutil"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	// "golang.org/x/crypto/bcrypt"
)

type Page struct {
	Title string
	Body  []byte
}

type User struct {
	Email           string
	Password        string
	ConfirmPassword string
	Token           string
	TokenExpiration string
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
		db.Prepare(`CREATE TABLE IF NOT EXISTS users (
			email VARCHAR(320) PRIMARY KEY, 
			hashedEmail CHAR(32), 
			hashedPassword CHAR(32),
			token CHAR(32),
			token_expiration CHAR(23)
			)`)
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
	rows, err := db.Query("SELECT email, hashedEmail, hashedPassword FROM users WHERE email = ?", u.Email)
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

// func CreateUser(w http.ResponseWriter, r *http.Request) {
//     user := &User{}
//     json.NewDecoder(r.Body).Decode(user)

//     pass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
//     if err != nil {
//         fmt.Println(err)
//         err := ErrorResponse{
//             Err: "Password Encryption failed",
//         }
//         json.NewEncoder(w).Encode(err)
//     }

//     user.Password = string(pass)

//     createdUser := db.Create(user)
// }

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, nil)
}

func newUserHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("new-user.html")
	t.Execute(w, nil)
}

func (u *User) validate() (string, string, error) {
	if time.Now().String() > u.TokenExpiration {
		return "", u.TokenExpiration, fmt.Errorf("Token Expired")
	}
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		return "", "", err
	}
	rows, err := db.Query("SELECT token, token_expiration FROM users WHERE email = ?", u.Email)
	if err != nil {
		return "", "", err
	}
	var oldToken string
	var tokenExpiration string
	for rows.Next() {
		rows.Scan(&oldToken, &tokenExpiration)
	}

	if oldToken == u.Token && tokenExpiration == u.TokenExpiration {
		newExpiration := time.Now().Add(time.Hour).String()
		tokenBytes := []byte(oldToken + newExpiration)
		tokenMD5 := md5.Sum(tokenBytes)
		newToken := hex.EncodeToString(tokenMD5[:])

		statement, err := db.Prepare("UPDATE users SET token_expiration = ?, token = ? WHERE email = ? and token = ?")
		if err != nil {
			return "", "", err
		}
		_, err = statement.Exec(newExpiration, newToken, u.Email, oldToken)
		fmt.Println(u.Email + ": " + newExpiration + " " + newToken)
		return newExpiration, newToken, err
	}
	return "", "", fmt.Errorf("Token and/or Token Expiration doesn't match")
}

func (u *User) login() (string, string, error) {
	passwordBytes := []byte(u.Password)
	passwordMD5 := md5.Sum(passwordBytes)

	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		return "", "", err
	}
	rows, err := db.Query("SELECT hashedPassword FROM users WHERE email = ?", u.Email)
	if err != nil {
		return "", "", err
	}
	var hashedPassword string
	for rows.Next() {
		rows.Scan(&hashedPassword)
	}
	if hashedPassword != hex.EncodeToString(passwordMD5[:]) {
		return "", "", fmt.Errorf("password incorrect")
	}
	expiration := time.Now().Add(time.Hour).String()

	tokenBytes := []byte(hashedPassword + expiration)
	tokenMD5 := md5.Sum(tokenBytes)
	token := hex.EncodeToString(tokenMD5[:])

	statement, err := db.Prepare("UPDATE users SET token_expiration = ?, token = ? WHERE email = ? and hashedPassword = ?")
	if err != nil {
		return "", "", err
	}
	_, err = statement.Exec(expiration, token, u.Email, hashedPassword)
	fmt.Println(u.Email + ": " + expiration + " " + token)
	return expiration, token, err
}

type Question struct {
	Email           string
	Token           string
	TokenExpiration string
	QID             int64
	Question        string
}

type Poll struct {
	Email           string
	Token           string
	TokenExpiration string
	PollID          int64
	Title           string
	Questions       []Question
}

type MenuData struct {
	Email           string
	Token           string
	TokenExpiration string
	Polls           []Poll
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		// logic part of log in
		u := &User{
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}
		fmt.Println("email:", u.Email)
		fmt.Println("password:", u.Password)

		token, expiration, err := u.login()
		if err != nil {
			fmt.Print(err)
		}
		data := MenuData{
			Email:           u.Email,
			Token:           token,
			TokenExpiration: expiration,
		}
		t, _ := template.ParseFiles("menu.html")
		t.Execute(w, data)
	}
}

func createPollHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	u := &User{
		Email:           r.FormValue("email"),
		Token:           r.FormValue("token"),
		TokenExpiration: r.FormValue("expiration"),
	}
	u.validate()
	if r.FormValue("new") != "" {
	}
	db, _ := sql.Open("sqlite3", "test.db")
	statement, _ :=
		db.Prepare(`CREATE TABLE IF NOT EXISTS polls (
			poll_id INTEGER PRIMARY KEY,
			email VARCHAR(320), 
			title TEXT
			)`)
	statement.Exec()
	statement, _ =
		db.Prepare(`CREATE TABLE IF NOT EXISTS questions (
			q_id INTEGER PRIMARY KEY,
			poll_id INTEGER,
			question TEXT
			)`)
	statement.Exec()
	poll := Poll{}
	if r.FormValue("poll") == "new" {
		statement, _ = db.Prepare("INSERT INTO polls (email) VALUES (?)")
		res, _ := statement.Exec(u.Email)
		poll.PollID, _ = res.LastInsertId()
		fmt.Printf("poll id: %v,", poll.PollID)
	}
	rows, _ := db.Query("SELECT title FROM polls WHERE poll_id = ?", poll.PollID)
	for rows.Next() {
		rows.Scan(&poll.Title)
	}
	rows, _ = db.Query("SELECT q_id, question FROM questions WHERE poll_id = ?", poll.PollID)
	for rows.Next() {
		q := Question{}
		rows.Scan(&q.QID, &q.Question)
		poll.Questions = append(poll.Questions, q)
	}
	t, _ := template.ParseFiles("create-poll.html")
	t.Execute(w, poll)
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
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/create", createPollHandler)

	fmt.Println("Listening on port 5050...")

	http.ListenAndServe(":5050", nil)
}
