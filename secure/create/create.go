package create

import (
	"fmt"
	"strconv"
	"html/template"

	"net/http"

	"github.com/maddatascience/simple-polling-web-app/models/poll"
	"github.com/maddatascience/simple-polling-web-app/models/user"
	"github.com/maddatascience/simple-polling-web-app/secure/menu"
)

func CreatePollHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	p := &poll.UserPoll{}

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
			m, err := menu.RenderMenu(u)
			if err != nil {
				fmt.Print(err)
			}
			t, err := template.ParseFiles("templates/menu.html")
			if err != nil {
				fmt.Print(err)
			}
			err = t.Execute(w, m)
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
