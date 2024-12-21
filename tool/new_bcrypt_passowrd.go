package main

import (
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := os.Getenv("PASSWORD")
	if password == "" {
		println("Please set password env PASSWORD")
		os.Exit(1)
	}
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		println("Generate bcrypt password failed", err)
		os.Exit(1)
	}
	println("Password: ", password)
	println("Bcrypt password: ", string(bcryptPassword))
}
