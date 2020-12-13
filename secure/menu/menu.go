package menu

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"github.com/maddatascience/simple-polling-web-app/models/user"

	"github.com/maddatascience/simple-polling-web-app/database"
	"github.com/maddatascience/simple-polling-web-app/models/poll"
)

type Menu struct {
	Polls []poll.Poll
}

func (m *Menu) Populate() error {
	db, err := database.InitDB(database.DataSourceName)
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
		m.Polls = append(m.Polls, p)
	}
	return err
}

type UserMenu struct {
	user.User
	Menu
}

func (m *UserMenu) Populate() error {
	db, err := database.InitDB(database.DataSourceName)
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

func RenderMenu(u *user.User) (*UserMenu, error) {
	fmt.Printf("\nuser: %v\n", u)
	menu := &UserMenu{
		User: *u,
	}
	err := fmt.Errorf("Neither Password nor Token provided")
	if u.Password != "" {
		err = u.Login()
	} else if u.Token != "" && u.TokenExpiration != "" {
		err = u.Validate()
	}
	if err != nil {
		return nil, err
	}
	fmt.Printf("%v", u)
	menu.Token = u.Token
	menu.TokenExpiration = u.TokenExpiration
	err = menu.Populate()
	return menu, err
}

func MenuHandler(w http.ResponseWriter, r *http.Request) {
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
	menu, err := RenderMenu(u)
	if err != nil {
		fmt.Print(err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
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
