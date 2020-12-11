package user

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/maddatascience/simple-polling-web-app/database"
)

type User struct {
	Email           string
	Password        string
	ConfirmPassword string
	Token           string
	TokenExpiration string
}

func (u *User) Save() error {
	if u.Password != u.ConfirmPassword {
		return errors.New("passwords do not match, try again.")
	}
	db, err := database.InitDB("test.db")
	if err != nil {
		return err
	}

	emailBytes := []byte(u.Email)
	passwordBytes := []byte(u.Password)
	emailMD5 := md5.Sum(emailBytes)
	passwordMD5 := md5.Sum(passwordBytes)
	hashedEmail := hex.EncodeToString(emailMD5[:])
	hashedPassword := hex.EncodeToString(passwordMD5[:])

	fmt.Println("hashedEmail:", hashedEmail)
	fmt.Println("hashedPassword:", hashedPassword)

	statement, err := db.Prepare("INSERT INTO users (email, hashedEmail, hashedPassword) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = statement.Exec(u.Email, hashedEmail, hashedPassword)
	if err != nil {
		return err
	}
	rows, err := db.Query("SELECT email, hashedEmail, hashedPassword FROM users WHERE email = ?", u.Email)
	if err != nil {
		return err
	}
	var email string
	for rows.Next() {
		err = rows.Scan(&email, &hashedEmail, &hashedPassword)
		if err != nil {
			return nil
		}
		fmt.Println(email + ": " + hashedEmail + " " + hashedPassword)
	}
	return nil
}

func (u *User) Validate() (string, string, error) {
	if time.Now().String() > u.TokenExpiration {
		return "", u.TokenExpiration, fmt.Errorf("Token Expired")
	}
	db, err := database.InitDB("test.db")
	if err != nil {
		return "", "", err
	}
	rows, err := db.Query("SELECT token, token_expiration FROM users WHERE email = ?", u.Email)
	if err != nil {
		return "", "", err
	}
	var oldToken string
	var tokenExpiration string
	for rows.Next() {
		err = rows.Scan(&oldToken, &tokenExpiration)
		if err != nil {
			return "", "", err
		}
	}

	if oldToken == u.Token && tokenExpiration == u.TokenExpiration {
		newExpiration := time.Now().Add(time.Hour).String()
		tokenBytes := []byte(oldToken + newExpiration)
		tokenMD5 := md5.Sum(tokenBytes)
		newToken := hex.EncodeToString(tokenMD5[:])

		statement, err := db.Prepare("UPDATE users SET token_expiration = ?, token = ? WHERE email = ? and token = ?")
		if err != nil {
			return "", "", err
		}
		_, err = statement.Exec(newExpiration, newToken, u.Email, oldToken)
		fmt.Println(u.Email + ": " + newExpiration + " " + newToken)
		return newExpiration, newToken, err
	}
	return "", "", fmt.Errorf("Token and/or Token Expiration doesn't match")
}

func (u *User) Login() (string, string, error) {
	passwordBytes := []byte(u.Password)
	passwordMD5 := md5.Sum(passwordBytes)

	db, err := database.InitDB("test.db")
	if err != nil {
		return "", "", err
	}
	rows, err := db.Query("SELECT hashedPassword FROM users WHERE email = ?", u.Email)
	if err != nil {
		return "", "", err
	}
	var hashedPassword string
	for rows.Next() {
		err = rows.Scan(&hashedPassword)
		if err != nil {
			return "", "", err
		}	
	}
	if hashedPassword != hex.EncodeToString(passwordMD5[:]) {
		return "", "", fmt.Errorf("password incorrect")
	}
	expiration := time.Now().Add(time.Hour).String()

	tokenBytes := []byte(hashedPassword + expiration)
	tokenMD5 := md5.Sum(tokenBytes)
	token := hex.EncodeToString(tokenMD5[:])

	statement, err := db.Prepare("UPDATE users SET token_expiration = ?, token = ? WHERE email = ? and hashedPassword = ?")
	if err != nil {
		return "", "", err
	}
	_, err = statement.Exec(expiration, token, u.Email, hashedPassword)
	fmt.Println(u.Email + ": " + expiration + " " + token)
	return expiration, token, err
}
