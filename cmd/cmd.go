package cmd

import (
	"strconv"

	"fmt"
	"html/template"

	"net/http"

	"github.com/maddatascience/simple-polling-web-app/database"
	"github.com/maddatascience/simple-polling-web-app/models/menu"
	"github.com/maddatascience/simple-polling-web-app/models/poll"
	"github.com/maddatascience/simple-polling-web-app/models/user"
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

func renderMenu(u *user.User) (*menu.UserMenu, error) {
	fmt.Printf("\nuser: %v\n", u)
	menu := &menu.UserMenu{
		User: *u,
	}
	err := fmt.Errorf("Neither Password nor Token provided")
	if u.Password != "" {
		err = u.Login()
	} else if u.Token != "" && u.TokenExpiration != "" {
		err = u.Validate()
	}
	if err != nil {
		return menu, err
	}
	fmt.Printf("%v", u)
	menu.Token = u.Token
	menu.TokenExpiration = u.TokenExpiration
	err = menu.Populate()
	return menu, err
}

func menuHandler(w http.ResponseWriter, r *http.Request) {
	println("parse form for menu")
	err := r.ParseForm()
	if err != nil {
		fmt.Print(err)
	}
	// logic part of log in
	u := &user.User{ // getting the old value from the form, not the updated version
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		Token:           r.FormValue("token"),
		TokenExpiration: r.FormValue("expiration"),
	}
	fmt.Println("email:", u.Email)
	menu, err := renderMenu(u)
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

func initPoll() (*poll.UserPoll, error) {
	p := &poll.UserPoll{}
	_, err := database.InitDB(database.DataSourceName)
	return p, err
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
	p, err := initPoll()
	if err != nil {
		fmt.Print(err)
	}

	if r.FormValue("poll-id") == "0" {
		err = p.New(u)
		if err != nil {
			fmt.Print(err)
		}
		fmt.Printf("poll id: %v,", p.PollID)
	} else {
		p.PollID, _ = strconv.ParseInt(r.FormValue("poll-id"), 10, 64)
	}

	err = p.Populate(u)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("\npoll %d has questions: %v\n", p.PollID, p.Questions)

	newQ := r.FormValue("new-question")

	if r.FormValue("page") == "create-poll" { // coming from create page
		// update title
		println("r.FormValue('page') == 'create-poll'")
		p.Title = r.FormValue("title")
		p.PollID, _ = strconv.ParseInt(r.FormValue("poll-id"), 10, 64)
		err = p.Update()
		if err != nil {
			fmt.Print(err)
		}

		// update questions
		for _, q := range p.Questions {
			fmt.Printf("found q: %v", q)
			if q.QuestionText != r.FormValue(strconv.FormatInt(q.QID, 10)) {
				q.QuestionText = r.FormValue(strconv.FormatInt(q.QID, 10))
				err = q.Update()
				if err != nil {
					fmt.Print(err)
				}
			}
		}

		// add new question
		if newQ != "" {
			newQID, err := p.NewQuestion(newQ)
			if err != nil {
				fmt.Print(err)
			}
			fmt.Printf("new question ID: %d\n", newQID)
		} else {
			println("return to menu")
			menu, err := renderMenu(u)
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
			return
		}
	}

	err = p.Populate(u)
	if err != nil {
		fmt.Print(err)
	}

	t, err := template.ParseFiles("templates/create-poll.html")
	if err != nil {
		fmt.Print(err)
	}
	err = t.Execute(w, p)
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

// func submitHandler(w http.ResponseWriter, r *http.Request) {

// }

func Execute(port string, dataSourceName string) error {
	database.DataSourceName = dataSourceName
	fmt.Printf("Setting database.DataSourceName to %s...\n", dataSourceName)
	http.HandleFunc("/", MainHandler)
	http.HandleFunc("/new-user", newUserHandler)
	http.HandleFunc("/save-user", saveUserHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/menu", menuHandler)
	http.HandleFunc("/create", createPollHandler)
	http.HandleFunc("/poll/", pollHandler)
	http.HandleFunc("/submit", MainHandler)

	fmt.Printf("Listening on port %s...\n", port)

	return http.ListenAndServe(":"+port, nil)
}
