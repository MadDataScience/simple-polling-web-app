package cmd

import (
	"database/sql"
	"strconv"

	"fmt"
	"html/template"

	"net/http"

	"github.com/maddatascience/simple-polling-web-app/database"
	"github.com/maddatascience/simple-polling-web-app/models/user"
	"github.com/maddatascience/simple-polling-web-app/models/poll"
	_ "github.com/mattn/go-sqlite3"
)

type Page struct {
	Title string
	Body  []byte
}

type Public struct {
	Polls []poll.Poll
}

func (pub *Public) populate() error {
	db, err := database.InitDB("test.db")
	if err != nil {
		return err
	}
	rows, err := db.Query("SELECT poll_id, title FROM polls")
	if err != nil {
		return err
	}
	p := poll.Poll{}
	var title sql.NullString
	for rows.Next() {
		err = rows.Scan(&p.PollID, &title)
		if err != nil {
			return err
		}
		if title.Valid {
			p.Title = title.String
		}
		pub.Polls = append(pub.Polls, p)
	}
	return err
}

type MenuData struct {
	Email           string
	Token           string
	TokenExpiration string
	Polls           []poll.Poll
}

func (m *MenuData) populate() error {
	db, err := database.InitDB("test.db")
	if err != nil {
		return err
	}
	rows, err := db.Query("SELECT poll_id, title FROM polls WHERE email = ?", m.Email)
	if err != nil {
		return err
	}
	p := poll.Poll{}
	var title sql.NullString
	for rows.Next() {
		err = rows.Scan(&p.PollID, &title)
		if err != nil {
			return err
		}
		if title.Valid {
			p.Title = title.String
		}
		m.Polls = append(m.Polls, p)
	}
	return err
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	p := &Public{}
	err := p.populate()
	if err != nil {
		fmt.Print(err)
	}
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Print(err)
	}
	err = t.Execute(w, p)
	if err != nil {
		fmt.Print(err)
	}
}

func newUserHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/new-user.html")
	t.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/login.html")
	t.Execute(w, nil)
}

func menuHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// logic part of log in
	u := &user.User{
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
		token, expiration, err := u.Login()
		if err != nil {
			fmt.Print(err)
		}
		menu.Token = token
		menu.TokenExpiration = expiration
		t, _ := template.ParseFiles("templates/menu.html")
		t.Execute(w, menu)
	} else if u.Token != "" && u.TokenExpiration != "" {
		u.Validate()
		menu.Token = u.Token
		menu.TokenExpiration = u.TokenExpiration
		menu.populate()
		t, _ := template.ParseFiles("templates/menu.html")
		t.Execute(w, menu)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func initPoll() (*poll.Poll, error) {
	poll := &poll.Poll{}
	_, err := database.InitDB("test.db")
	return poll, err
}



func createPollHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	u := &user.User{
		Email:           r.FormValue("email"),
		Token:           r.FormValue("token"),
		TokenExpiration: r.FormValue("expiration"),
	}
	u.Validate()
	poll, err := initPoll()
	if err != nil {
		fmt.Print(err)
	}

	if r.FormValue("poll") == "new" {
		err = poll.New(u)
		if err != nil {
			fmt.Print(err)
		}
		fmt.Printf("poll id: %v,", poll.PollID)
	} else {
		poll.PollID, _ = strconv.ParseInt(r.FormValue("poll"), 10, 64)
	}

	poll.Populate(u)

	newQ := r.FormValue("new-question")

	if r.FormValue("page") == "create-poll" { // coming from create page
		// update title
		poll.Title = r.FormValue("title")
		poll.PollID, _ = strconv.ParseInt(r.FormValue("poll-id"), 10, 64)
		poll.Update()

		// update questions
		for _, q := range poll.Questions {
			if q.Question != r.FormValue(strconv.FormatInt(q.QID, 10)) {
				q.Question = r.FormValue(strconv.FormatInt(q.QID, 10))
				q.Update()
			}
		}

		// add new question
		if newQ != "" {
			newQID, err := poll.NewQuestion(newQ)
			if err != nil {
				fmt.Print(err)
			}
			fmt.Printf("new question ID: %d\n", newQID)
		} else {
			menuHandler(w, r)
			return
		}
	}

	poll.Populate(u)

	t, _ := template.ParseFiles("templates/create-poll.html")
	t.Execute(w, poll)
}

func saveUserHandler(w http.ResponseWriter, r *http.Request) {
	u := &user.User{
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirm-password"),
	}
	err := u.Save()
	if err != nil {
		fmt.Print(err)
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func Execute() error {
	http.HandleFunc("/", MainHandler)
	http.HandleFunc("/new-user", newUserHandler)
	http.HandleFunc("/save-user", saveUserHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/menu", menuHandler)
	http.HandleFunc("/create", createPollHandler)

	fmt.Println("Listening on port 5050...")

	return http.ListenAndServe(":5050", nil)
}
