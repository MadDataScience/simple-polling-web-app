package cmd

import (
	"database/sql"
	"strconv"

	"fmt"
	"html/template"

	"net/http"

	"github.com/maddatascience/simple-polling-web-app/database"
	"github.com/maddatascience/simple-polling-web-app/models/user"
	_ "github.com/mattn/go-sqlite3"
)

type Page struct {
	Title string
	Body  []byte
}

type Question struct {
	Email           string
	Token           string
	TokenExpiration string
	QID             int64
	Question        string
}

func (q *Question) update() error {
	db, err := database.InitDB("test.db")
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
	db, err := database.InitDB("test.db")
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
	db, err := database.InitDB("test.db")
	if err != nil {
		return err
	}
	rows, err := db.Query("SELECT poll_id, title FROM polls WHERE email = ?", m.Email)
	if err != nil {
		return err
	}
	p := Poll{}
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

type Public struct {
	Polls []Poll
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
	p := Poll{}
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

func initPoll() (*Poll, error) {
	poll := &Poll{}
	_, err := database.InitDB("test.db")
	return poll, err
}

func (p *Poll) new(u *user.User) error {
	if u.Email == "" {
		return fmt.Errorf("email missing")
	}
	p.Email = u.Email
	p.Token = u.Token
	p.TokenExpiration = u.TokenExpiration

	db, err := database.InitDB("test.db")
	if err != nil {
		return err
	}
	statement, err := db.Prepare("INSERT INTO polls (email) VALUES (?)")
	if err != nil {
		return err
	}
	res, err := statement.Exec(u.Email)
	if err != nil {
		return err
	}
	p.PollID, err = res.LastInsertId()
	return err
}

func (p *Poll) populate(u *user.User) error {
	if p.PollID == 0 {
		return fmt.Errorf("poll id missing")
	}
	p.Email = u.Email
	p.Token = u.Token
	p.TokenExpiration = u.TokenExpiration

	db, err := database.InitDB("test.db")
	if err != nil {
		return err
	}
	rows, err := db.Query("SELECT title FROM polls WHERE poll_id = ?", p.PollID)
	if err != nil {
		return err
	}
	var title sql.NullString
	for rows.Next() {
		rows.Scan(&title)
	}
	if title.Valid {
		p.Title = title.String
	}
	rows, err = db.Query("SELECT q_id, question FROM questions WHERE poll_id = ?", p.PollID)
	if err != nil {
		return err
	}
	p.Questions = []Question{} // reset question array
	q := Question{}
	for rows.Next() {
		err = rows.Scan(&q.QID, &q.Question)
		if err != nil {
			return err
		}
		p.Questions = append(p.Questions, q)
	}
	return err
}

func (poll *Poll) newQuestion(question string) (int64, error) {
	if poll.PollID == 0 {
		return 0, fmt.Errorf("poll id missing")
	}
	db, err := database.InitDB("test.db")
	if err != nil {
		return 0, err
	}
	statement, err := db.Prepare("INSERT INTO questions (poll_id, question) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	res, err := statement.Exec(poll.PollID, question)
	if err != nil {
		return 0, err
	}
	newQID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	newQ := &Question{
		QID:      newQID,
		Question: question,
	}
	poll.Questions = append(poll.Questions, *newQ)
	return newQ.QID, err
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
		err = poll.new(u)
		if err != nil {
			fmt.Print(err)
		}
		fmt.Printf("poll id: %v,", poll.PollID)
	} else {
		poll.PollID, _ = strconv.ParseInt(r.FormValue("poll"), 10, 64)
	}

	poll.populate(u)

	newQ := r.FormValue("new-question")

	if r.FormValue("page") == "create-poll" { // coming from create page
		// update title
		poll.Title = r.FormValue("title")
		poll.PollID, _ = strconv.ParseInt(r.FormValue("poll-id"), 10, 64)
		poll.update()

		// update questions
		for _, q := range poll.Questions {
			if q.Question != r.FormValue(strconv.FormatInt(q.QID, 10)) {
				q.Question = r.FormValue(strconv.FormatInt(q.QID, 10))
				q.update()
			}
		}

		// add new question
		if newQ != "" {
			newQID, err := poll.newQuestion(newQ)
			if err != nil {
				fmt.Print(err)
			}
			fmt.Printf("new question ID: %d\n", newQID)
		} else {
			menuHandler(w, r)
			return
		}
	}

	poll.populate(u)

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

func Execute() {
	// err := database.InitDB("test.db")

	http.HandleFunc("/", MainHandler)
	http.HandleFunc("/new-user", newUserHandler)
	http.HandleFunc("/save-user", saveUserHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/menu", menuHandler)
	http.HandleFunc("/create", createPollHandler)

	fmt.Println("Listening on port 5050...")

	http.ListenAndServe(":5050", nil)
}
