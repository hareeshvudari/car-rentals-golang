package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// Struct to represent the user
type User struct {
	ID           int    `json:"id"`
	MobileNumber string `json:"mobile_number"`
}

// Struct to represent the OTP
type OTP struct {
	Value string `json:"value"`
}

// Generate a random OTP of length 6
func generateOTP() string {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(900000) + 100000
	return fmt.Sprintf("%d", otp)
}

// Send the OTP to the user's mobile number
func sendOTP(mobileNumber string, otp string) {
	// Implement your SMS API logic here
	fmt.Printf("OTP sent to %s: %s\n", mobileNumber, otp)
}

// Verify if the entered OTP is correct
func verifyOTP(enteredOTP string, generatedOTP string) bool {
	return enteredOTP == generatedOTP
}

// Login with OTP handler
func loginWithOTP(w http.ResponseWriter, r *http.Request) {
	// Parse the request body into a User struct
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the user exists in the database
	// Assume that the user is already registered and stored in the database
	// and retrieve the user's ID
	// userID := 1

	// Generate an OTP
	otp := generateOTP()

	// Send the OTP to the user's mobile number
	sendOTP(user.MobileNumber, otp)

	// Return the OTP in the response body
	otpResponse := OTP{Value: otp}
	json.NewEncoder(w).Encode(otpResponse)

	// Store the OTP in the database with the user ID and timestamp
	// Assume that the OTP and its metadata are stored in the database
}

// Verify OTP handler
func verifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body into an OTP struct
	var otp OTP
	err := json.NewDecoder(r.Body).Decode(&otp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve the user's ID from the request URL parameters
	params := mux.Vars(r)
	userID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve the latest OTP for the user from the database
	// Assume that the OTP and its metadata are stored in the database
	// and retrieve the OTP and its timestamp
	latestOTP := "123456"
	latestOTPTimestamp := time.Now()

	// Verify if the entered OTP is correct
	if verifyOTP(otp.Value, latestOTP) {
		// Check if the OTP is expired
		expiryTime := latestOTPTimestamp.Add(time.Minute * 5) // Assume OTP is valid for 5 minutes
		if time.Now().After(expiryTime) {
			http.Error(w, "OTP has expired. Please request a new one.", http.StatusUnauthorized)
			return
		}

		// Return a success response with the user ID
		jsonResponse := map[string]int{"user_id": userID}
		json.NewEncoder(w).Encode(jsonResponse)
		// Delete the OTP from the database
		// Assume that the OTP and its metadata are deleted from the database
	} else {
		http.Error(w, "Invalid OTP. Please try again.", http.StatusUnauthorized)
		return
	}
}

func main() {
	// Create a new router
	router := mux.NewRouter()

	// Define the login with OTP endpoint
	router.HandleFunc("/login", loginWithOTP).Methods("POST")

	// Define the verify OTP endpoint with a user ID parameter
	router.HandleFunc("/verify/{id:[0-9]+}", verifyOTPHandler).Methods("POST")

	// Start the server
	fmt.Println("Server listening on port 8080")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println(err)
	}
}
