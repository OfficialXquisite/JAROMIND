// hash_password.go
package main

import (
	"fmt"
	"log"
	
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Your desired admin password
	password := "Admin123" // CHANGE THIS TO YOUR SECURE PASSWORD
	
	// Generate hashed password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Error:", err)
	}
	
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Hashed:   %s\n", string(hashedPassword))
	
	// Also show MongoDB insert command
	fmt.Println("\n📋 MongoDB Insert Command:")
	fmt.Printf(`
db.admins.insertOne({
  email: "admin@jaromind.com",
  password: "%s",
  name: "Super Admin",
  createdAt: new Date(),
  isActive: true
})
`, string(hashedPassword))
}