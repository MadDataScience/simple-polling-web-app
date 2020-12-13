package poll

import (
	"database/sql"
	"fmt"

	"github.com/maddatascience/simple-polling-web-app/database"
	"github.com/maddatascience/simple-polling-web-app/models/question"
	"github.com/maddatascience/simple-polling-web-app/models/user"
)

type Poll struct {
	PollID    int64
	Title     string
	Questions []question.Question
}

type UserPoll struct {
	user.User
	Poll
}

func (p *UserPoll) Update() error {
	db, err := database.InitDB(database.DataSourceName)
	if err != nil {
		return err
	}
	statement, err := db.Prepare("UPDATE polls SET title = ? WHERE poll_id = ?")
	if err != nil {
		return err
	}
	_, err = statement.Exec(p.Title, p.PollID)
	// fmt.Printf("Poll %d: %s\n", p.PollID, p.Title)
	return err
}

func (p *UserPoll) New(u *user.User) error {
	if u.Email == "" {
		return fmt.Errorf("email missing")
	}
	p.Email = u.Email
	p.Token = u.Token
	p.TokenExpiration = u.TokenExpiration

	db, err := database.InitDB(database.DataSourceName)
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

func (p *Poll) Populate() error {
	if p.PollID == 0 {
		return fmt.Errorf("poll id missing")
	}
	db, err := database.InitDB(database.DataSourceName)
	if err != nil {
		return err
	}
	rows, err := db.Query("SELECT title FROM polls WHERE poll_id = ?", p.PollID)
	if err != nil {
		return err
	}
	var title sql.NullString
	for rows.Next() {
		err = rows.Scan(&title)
		if err != nil {
			return err
		}
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
		err = rows.Scan(&q.QID, &q.QuestionText)
		if err != nil {
			return err
		}
		p.Questions = append(p.Questions, q)
	}
	return err
}

func (p *UserPoll) Populate(u *user.User) error {
	p.Email = u.Email
	p.Token = u.Token
	p.TokenExpiration = u.TokenExpiration
	return p.Poll.Populate()
}

func (p *UserPoll) NewQuestion(questionText string) (int64, error) {
	if p.PollID == 0 {
		return 0, fmt.Errorf("poll id missing")
	}
	db, err := database.InitDB(database.DataSourceName)
	if err != nil {
		return 0, err
	}
	statement, err := db.Prepare("INSERT INTO questions (poll_id, question) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	res, err := statement.Exec(p.PollID, questionText)
	if err != nil {
		return 0, err
	}
	newQID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	newQ := &question.Question{
		QID:          newQID,
		QuestionText: questionText,
	}
	p.Questions = append(p.Questions, *newQ)
	return newQ.QID, err
}

func RetrievePoll(pollID int64) (*Poll, error) {
	p := &Poll{
		PollID: pollID,
	}
	err := p.Populate()
	return p, err
}