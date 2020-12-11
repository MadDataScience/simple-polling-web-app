package question

import (
	"fmt"
	"github.com/maddatascience/simple-polling-web-app/database"
)

type Question struct {
	Email           string
	Token           string
	TokenExpiration string
	QID             int64
	Question        string
}

func (q *Question) Update() error {
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
