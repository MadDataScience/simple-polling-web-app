package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"strconv"
	"time"

	// "encoding/json"
	"errors"
	"fmt"
	"html/template"

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

	statement, err := db.Prepare("INSERT INTO users (email, hashedEmail, hashedPassword) VALUES (?, ?, ?)")
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

func (q *Question) update() error {
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		return err
	}
	statement, err := db.Prepare("UPDATE questions SET question = ? WHERE q_id = ?")
	if err != nil {
		return err
	}
	_, err = statement.Exec(q.Question, q.QID)
	fmt.Printf("Question %d: %s\n", q.QID, q.Question)
	return err
}

type Poll struct {
	Email           string
	Token           string
	TokenExpiration string
	PollID          int64
	Title           string
	Questions       []Question
}

func (p *Poll) update() error {
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		return err
	}
	statement, err := db.Prepare("UPDATE polls SET title = ? WHERE poll_id = ?")
	if err != nil {
		return err
	}
	_, err = statement.Exec(p.Title, p.PollID)
	fmt.Printf("Poll %d: %s\n", p.PollID, p.Title)
	return err
}

type MenuData struct {
	Email           string
	Token           string
	TokenExpiration string
	Polls           []Poll
}

func (m *MenuData) populate() error {
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		return err
	}
	rows, err := db.Query("SELECT poll_id, title FROM polls WHERE email = ?", m.Email)
	if err != nil {
		return err
	}
	p := Poll{}
	for rows.Next() {
		err = rows.Scan(&p.PollID, &p.Title)
		if err != nil {
			return err
		}
		m.Polls = append(m.Polls, p)
	}
	return err
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, nil)
}

func newUserHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("new-user.html")
	t.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("login.html")
	t.Execute(w, nil)
}

func menuHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// logic part of log in
	u := &User{
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		Token:           r.FormValue("token"),
		TokenExpiration: r.FormValue("expiration"),
	}
	fmt.Println("email:", u.Email)
	menu := MenuData{
		Email: u.Email,
	}

	if u.Password != "" {
		token, expiration, err := u.login()
		if err != nil {
			fmt.Print(err)
		}
		menu.Token = token
		menu.TokenExpiration = expiration
		t, _ := template.ParseFiles("menu.html")
		t.Execute(w, menu)
	} else if u.Token != "" && u.TokenExpiration != "" {
		u.validate()
		menu.Token = u.Token
		menu.TokenExpiration = u.TokenExpiration
		menu.populate()
		t, _ := template.ParseFiles("menu.html")
		t.Execute(w, menu)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
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
	} else {
		poll.PollID, _ = strconv.ParseInt(r.FormValue("poll"), 10, 64)
	}

	rows, _ := db.Query("SELECT q_id, question FROM questions WHERE poll_id = ?", poll.PollID)
	q := Question{}
	for rows.Next() {
		rows.Scan(&q.QID, &q.Question)
		poll.Questions = append(poll.Questions, q)
	}

	newQ := r.FormValue("new-question")

	if r.FormValue("page") == "create-poll" { // coming from create page
		// update title
		poll.Title = r.FormValue("title")
		poll.PollID, _ = strconv.ParseInt(r.FormValue("poll-id"), 10, 64)
		poll.update()

		// update questions
		for _, q = range poll.Questions {
			if q.Question != r.FormValue(strconv.FormatInt(q.QID, 10)) {
				q.Question = r.FormValue(strconv.FormatInt(q.QID, 10))
				q.update()
			}
		}

		// add new question
		if newQ != "" {
			statement, _ = db.Prepare("INSERT INTO questions (poll_id, question) VALUES (?, ?)")
			res, _ := statement.Exec(poll.PollID, newQ)
			newQID, _ := res.LastInsertId()
			fmt.Printf("new question ID: %d\n", newQID)
		} else {
			menuHandler(w, r)
			return
		}
	}

	refreshPoll := Poll{
		PollID:          poll.PollID,
		Email:           u.Email,
		Token:           u.Token,
		TokenExpiration: u.TokenExpiration,
	}

	rows, _ = db.Query("SELECT title FROM polls WHERE poll_id = ?", refreshPoll.PollID)
	for rows.Next() {
		rows.Scan(&refreshPoll.Title)
	}

	rows, _ = db.Query("SELECT q_id, question FROM questions WHERE poll_id = ?", refreshPoll.PollID)
	for rows.Next() {
		q = Question{}
		rows.Scan(&q.QID, &q.Question)
		refreshPoll.Questions = append(refreshPoll.Questions, q)
	}

	t, _ := template.ParseFiles("create-poll.html")
	t.Execute(w, refreshPoll)
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
	db, _ := sql.Open("sqlite3", "test.db")
	statement, err :=
		db.Prepare(`CREATE TABLE IF NOT EXISTS users (
			email VARCHAR(320) PRIMARY KEY, 
			hashedEmail CHAR(32), 
			hashedPassword CHAR(32),
			token CHAR(32),
			token_expiration CHAR(23)
			)`)
	if err != nil {
		fmt.Print(err)
	}
	_, err = statement.Exec()
	if err != nil {
		fmt.Print(err)
	}
	http.HandleFunc("/", MainHandler)
	http.HandleFunc("/new-user", newUserHandler)
	http.HandleFunc("/save-user", saveUserHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/menu", menuHandler)
	http.HandleFunc("/create", createPollHandler)

	fmt.Println("Listening on port 5050...")

	http.ListenAndServe(":5050", nil)
}
