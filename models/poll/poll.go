package poll

import (
	"fmt"
	"database/sql"
	"github.com/maddatascience/simple-polling-web-app/database"
	"github.com/maddatascience/simple-polling-web-app/models/question"
	"github.com/maddatascience/simple-polling-web-app/models/user"
)

type Poll struct {
	Email           string
	Token           string
	TokenExpiration string
	PollID          int64
	Title           string
	Questions       []question.Question
}

func (p *Poll) Update() error {
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


func (p *Poll) New(u *user.User) error {
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

func (p *Poll) Populate(u *user.User) error {
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
	p.Questions = []question.Question{} // reset question array
	q := question.Question{}
	for rows.Next() {
		err = rows.Scan(&q.QID, &q.Question)
		if err != nil {
			return err
		}
		p.Questions = append(p.Questions, q)
	}
	return err
}

func (poll *Poll) NewQuestion(questionText string) (int64, error) {
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
	res, err := statement.Exec(poll.PollID, questionText)
	if err != nil {
		return 0, err
	}
	newQID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	newQ := &question.Question{
		QID:      newQID,
		Question: questionText,
	}
	poll.Questions = append(poll.Questions, *newQ)
	return newQ.QID, err
}
