package cmd

import (
	"strconv"

	"fmt"
	"html/template"

	"net/http"

	"github.com/maddatascience/simple-polling-web-app/database"
	"github.com/maddatascience/simple-polling-web-app/models/poll"
	"github.com/maddatascience/simple-polling-web-app/models/user"
	"github.com/maddatascience/simple-polling-web-app/secure/create"
	"github.com/maddatascience/simple-polling-web-app/secure/menu"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	p := &menu.Menu{}
	err := p.Populate()
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

func pollHandler(w http.ResponseWriter, r *http.Request) {
	p := &poll.Poll{}
	pollIDstr := r.URL.Path[len("/poll/"):]
	p.PollID, _ = strconv.ParseInt(pollIDstr, 10, 64)
	err := p.Populate()
	if err != nil {
		fmt.Print(err)
	}
	t, err := template.ParseFiles("templates/poll.html")
	if err != nil {
		fmt.Print(err)
	}
	err = t.Execute(w, p)
	if err != nil {
		fmt.Print(err)
	}
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Print(err)
	}
	pollID, err := strconv.ParseInt(r.FormValue("poll-id"), 10, 64)
	if err != nil {
		fmt.Print(err)
	}
	p, err := poll.RetrievePoll(pollID)
	if err != nil {
		fmt.Print(err)
	}
	for _, q := range p.Questions {
		fmt.Printf("found q: %v", q.QID)
		answer := r.FormValue(strconv.FormatInt(q.QID, 10))
		fmt.Printf("found q: %v - answer: %s", q.QID, answer)
		answerInt := 0
		if answer != "" {
			answerInt = 1
		}
		err = q.Answer(answerInt)
		if err != nil {
			fmt.Print(err)
		}
	}
	MainHandler(w, r)
}

func Execute(port string, dataSourceName string) error {
	database.DataSourceName = dataSourceName
	fmt.Printf("Setting database.DataSourceName to %s...\n", dataSourceName)
	http.HandleFunc("/", MainHandler)
	http.HandleFunc("/new-user", newUserHandler)
	http.HandleFunc("/save-user", saveUserHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/menu", menu.MenuHandler)
	http.HandleFunc("/create", create.CreatePollHandler)
	http.HandleFunc("/poll/", pollHandler)
	http.HandleFunc("/submit", submitHandler)

	fmt.Printf("Listening on port %s...\n", port)

	return http.ListenAndServe(":"+port, nil)
}
