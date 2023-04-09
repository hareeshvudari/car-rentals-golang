package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type User struct {
	ID       string `json:"id" bson:"_id,omitempty"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ErrorResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

type SuccessResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

var (
	client     *mongo.Client
	collection *mongo.Collection
)

func main() {
	// Set up MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	collection = client.Database("rental_cars").Collection("users")

	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/signin", signinHandler)
	http.HandleFunc("/forgot-password", forgotPasswordHandler)
	http.HandleFunc("/updatepassword", updatePassword)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user already exists in the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	existingUser := collection.FindOne(ctx, map[string]string{"email": user.Email})
	if existingUser.Err() == nil {
		// User already exists in the database, return an error message
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "User already exists", StatusCode: http.StatusConflict})
		return
	}

	// Insert new user into the database
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: "User added successfully", StatusCode: http.StatusOK})
}
func signinHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user exists in the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	existingUser := collection.FindOne(ctx, bson.M{"email": user.Email, "password": user.Password})
	if existingUser.Err() != nil {
		// User does not exist in the database, return an error message
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid email or password", StatusCode: http.StatusUnauthorized})
		return
	}

	// Send response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: "Login successful", StatusCode: http.StatusOK})
}

func forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user exists in the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	existingUser := collection.FindOne(ctx, bson.M{"email": user.Email})
	if existingUser.Err() != nil {
		// User does not exist in the database, return an error message
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "User does not exist"})
		return
	}

	// Generate a unique token
	token := generateToken()

	// Store the token in the database for the user
	_, err = collection.UpdateOne(ctx, bson.M{"email": user.Email}, bson.M{"$set": bson.M{"resetToken": token}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send password reset link to user's email
	sendPasswordResetLink(user.Email, token)

	// Send response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{StatusCode: http.StatusOK, Message: "Password reset link sent successfully"})
}

func generateToken() string {
	// Generate 32 bytes of random data
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		// Handle error
	}

	// Encode the random data as a base64 string
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	return token
}

func sendPasswordResetLink(email string, token string) {

	from := "vduari1043@gmail.com"
	password := "Saha@1043"
	to := []string{email}

	url := fmt.Sprintf("https://example.com/reset-password?token=%s", token)
	msg := fmt.Sprintf("To reset your password, please click on the following link: %s", url)

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")

	err := smtp.SendMail("smtp.gmail.com:587", auth, from, to, []byte(msg))
	if err != nil {
		log.Println(err)
	}
}

func updatePassword(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request body
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Connect to MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if user exists
	collection := client.Database("rental_cars").Collection("users")
	filter := bson.M{"email": user.Email}
	var existingUser User
	err = collection.FindOne(ctx, filter).Decode(&existingUser)
	if err != nil {
		http.Error(w, "User does not exist", http.StatusNotFound)
		return
	}

	// Update user's password
	existingUser.Password = user.Password
	update := bson.M{"$set": bson.M{"password": existingUser.Password}}
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Password updated successfully")
}
