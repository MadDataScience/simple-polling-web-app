package menu

import (
	"database/sql"

	"github.com/maddatascience/simple-polling-web-app/database"
	"github.com/maddatascience/simple-polling-web-app/models/poll"
	"github.com/maddatascience/simple-polling-web-app/models/user"
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
