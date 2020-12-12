package cmd

import (
	"database/sql"
	"strconv"

	"fmt"
	"html/template"

	"net/http"

	"github.com/maddatascience/simple-polling-web-app/database"
	"github.com/maddatascience/simple-polling-web-app/models/poll"
	"github.com/maddatascience/simple-polling-web-app/models/user"
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
	t, err := template.ParseFiles("templates/new-user.html")
	if err != nil {
		fmt.Print(err)
	}
	err = t.Execute(w, nil)
	if err != nil {
		fmt.Print(err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/login.html")
	if err != nil {
		fmt.Print(err)
	}
	err = t.Execute(w, nil)
	if err != nil {
		fmt.Print(err)
	}
}

func menuHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Print(err)
	}
	// logic part of log in
	u := &user.User{
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		Token:           r.FormValue("token"),
		TokenExpiration: r.FormValue("expiration"),
	}
	fmt.Println("email:", u.Email)
	fmt.Printf("\nuser: %v\n", u)
	menu := MenuData{
		Email: u.Email,
	}

	if u.Password != "" {
		err = u.Login()
	} else if u.Token != "" && u.TokenExpiration != "" {
		err = u.Validate()
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("%v", u)
	menu.Token = u.Token
	menu.TokenExpiration = u.TokenExpiration
	err = menu.populate()
	if err != nil {
		fmt.Print(err)
	}
	t, err := template.ParseFiles("templates/menu.html")
	if err != nil {
		fmt.Print(err)
	}
	err = t.Execute(w, menu)
	if err != nil {
		fmt.Print(err)
	}
}

func initPoll() (*poll.Poll, error) {
	poll := &poll.Poll{}
	_, err := database.InitDB("test.db")
	return poll, err
}

func createPollHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Print(err)
	}
	u := &user.User{
		Email:           r.FormValue("email"),
		Token:           r.FormValue("token"),
		TokenExpiration: r.FormValue("expiration"),
	}
	err = u.Validate()
	if err != nil {
		fmt.Print(err)
	}
	poll, err := initPoll()
	if err != nil {
		fmt.Print(err)
	}

	if r.FormValue("poll-id") == "0" {
		err = poll.New(u)
		if err != nil {
			fmt.Print(err)
		}
		fmt.Printf("poll id: %v,", poll.PollID)
	} else {
		poll.PollID, _ = strconv.ParseInt(r.FormValue("poll-id"), 10, 64)
	}

	err = poll.Populate(u)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("\npoll %d has questions: %v\n", poll.PollID, poll.Questions)

	newQ := r.FormValue("new-question")

	if r.FormValue("page") == "create-poll" { // coming from create page
		// update title
		println("r.FormValue('page') == 'create-poll'")
		poll.Title = r.FormValue("title")
		poll.PollID, _ = strconv.ParseInt(r.FormValue("poll-id"), 10, 64)
		err = poll.Update()
		if err != nil {
			fmt.Print(err)
		}

		// update questions
		for _, q := range poll.Questions {
			fmt.Printf("found q: %v", q)
			if q.Question != r.FormValue(strconv.FormatInt(q.QID, 10)) {
				q.Question = r.FormValue(strconv.FormatInt(q.QID, 10))
				err = q.Update()
				if err != nil {
					fmt.Print(err)
				}
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

	err = poll.Populate(u)
	if err != nil {
		fmt.Print(err)
	}

	t, err := template.ParseFiles("templates/create-poll.html")
	if err != nil {
		fmt.Print(err)
	}
	err = t.Execute(w, poll)
	if err != nil {
		fmt.Print(err)
	}
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
