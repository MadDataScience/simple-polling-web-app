package question

import (
	"fmt"

	"github.com/maddatascience/simple-polling-web-app/database"
)

type Question struct {
	QID          int64
	QuestionText string
}

func (q *Question) Update() error {
	db, err := database.InitDB(database.DataSourceName)
	if err != nil {
		return err
	}
	statement, err := db.Prepare("UPDATE questions SET question = ? WHERE q_id = ?")
	if err != nil {
		return err
	}
	_, err = statement.Exec(q.QuestionText, q.QID)
	// fmt.Printf("Question %d: %s\n", q.QID, q.QuestionText)
	return err
}

func (q *Question) Answer(a int) error {
	if q.QID == 0 {
		return fmt.Errorf("question id missing")
	}
	db, err := database.InitDB(database.DataSourceName)
	if err != nil {
		return err
	}
	statement, err := db.Prepare("INSERT INTO answers (q_id, answerInt) VALUES (?, ?)")
	if err != nil {
		return err
	}
	_, err = statement.Exec(q.QID, a)
	return err
}
