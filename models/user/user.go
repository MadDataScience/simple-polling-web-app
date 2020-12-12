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

func (u *User) Login() error {
	println("login")
	passwordBytes := []byte(u.Password)
	passwordMD5 := md5.Sum(passwordBytes)

	db, err := database.InitDB("test.db")
	if err != nil {
		return err
	}
	rows, err := db.Query("SELECT hashedPassword FROM users WHERE email = ?", u.Email)
	if err != nil {
		return err
	}
	var hashedPassword string
	for rows.Next() {
		err = rows.Scan(&hashedPassword)
		if err != nil {
			return err
		}
	}
	if hashedPassword != hex.EncodeToString(passwordMD5[:]) {
		return fmt.Errorf("password incorrect")
	}
	u.TokenExpiration = time.Now().Add(time.Hour).String()

	tokenBytes := []byte(hashedPassword + u.TokenExpiration)
	tokenMD5 := md5.Sum(tokenBytes)
	u.Token = hex.EncodeToString(tokenMD5[:])

	statement, err := db.Prepare("UPDATE users SET token_expiration = ?, token = ? WHERE email = ? and hashedPassword = ?")
	if err != nil {
		return err
	}
	_, err = statement.Exec(u.TokenExpiration, u.Token, u.Email, hashedPassword)
	fmt.Println(u.Email + ": " + u.TokenExpiration + " " + u.Token)
	return err
}

func (u *User) Validate() error {
	println("validate")
	if time.Now().String() > u.TokenExpiration {
		return fmt.Errorf("\nToken Expired: %s > %s\n", time.Now().String(), u.TokenExpiration)
	}
	db, err := database.InitDB("test.db")
	if err != nil {
		return err
	}
	rows, err := db.Query("SELECT token, token_expiration FROM users WHERE email = ?", u.Email)
	if err != nil {
		return err
	}
	var oldToken string
	var tokenExpiration string
	for rows.Next() {
		err = rows.Scan(&oldToken, &tokenExpiration)
		if err != nil {
			return err
		}
	}
	fmt.Printf("\ncomparing %v to \n\t(%s --- %s)", u, oldToken, tokenExpiration)
	if oldToken == u.Token && tokenExpiration == u.TokenExpiration {
		fmt.Printf("\n%s and %s match, updating...\n", oldToken, tokenExpiration)
		newExpiration := time.Now().Add(time.Hour).String()
		tokenBytes := []byte(oldToken + u.TokenExpiration)
		tokenMD5 := md5.Sum(tokenBytes)
		newToken := hex.EncodeToString(tokenMD5[:])

		statement, err := db.Prepare("UPDATE users SET token_expiration = ?, token = ? WHERE email = ? and token = ?")
		if err != nil {
			return err
		}
		res, err := statement.Exec(newExpiration, newToken, u.Email, oldToken)
		fmt.Printf("\ttransaction result: %v\n", res)
		fmt.Printf("\t%s: %s <- %s; %s <- %s\n", u.Email, u.TokenExpiration, newExpiration, u.Token, newToken)
		u.TokenExpiration = newExpiration
		u.Token = newToken
		return err
	}
	return fmt.Errorf("Token and/or Token Expiration doesn't match")
}
