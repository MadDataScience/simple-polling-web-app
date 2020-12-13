package auth

import (
	"fmt"
	"html/template"

	"net/http"

	"github.com/maddatascience/simple-polling-web-app/models/user"
)

type AuthData struct {
	Warn string
}

func NewUserHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Print(err)
	}
	data := AuthData{
		Warn: r.FormValue("error"),
	}
	t, err := template.ParseFiles("templates/new-user.html")
	if err != nil {
		fmt.Print(err)
	}
	err = t.Execute(w, data)
	if err != nil {
		fmt.Print(err)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Print(err)
	}
	data := AuthData{
		Warn: r.FormValue("error"),
	}
	t, err := template.ParseFiles("templates/login.html")
	if err != nil {
		fmt.Print(err)
	}
	err = t.Execute(w, data)
	if err != nil {
		fmt.Print(err)
	}
}

func SaveUserHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Print(err)
	}
	u := &user.User{
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirm-password"),
	}
	err = u.Save()
	if err != nil {
		println(err.Error())
		r.Form.Add("error", err.Error())
		NewUserHandler(w, r)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}
