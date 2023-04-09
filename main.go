package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/gorilla/mux"
)

type Car struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Brand       string             `json:"brand,omitempty" bson:"brand,omitempty"`
	Model       string             `json:"model,omitempty" bson:"model,omitempty"`
	Year        int                `json:"year,omitempty" bson:"year,omitempty"`
	Color       string             `json:"color,omitempty" bson:"color,omitempty"`
	DailyRate   float64            `json:"daily_rate,omitempty" bson:"daily_rate,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	UpdatedAt   time.Time
}

var client *mongo.Client

func main() {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	var err error
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Ping the MongoDB server to check the connection
	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	// Create a new HTTP router
	router := mux.NewRouter()

	// Define the HTTP routes
	router.HandleFunc("/cars", getCars).Methods("GET")
	router.HandleFunc("/cars", createCar).Methods("POST")
	router.HandleFunc("/cars/{id}", getCar).Methods("GET")
	router.HandleFunc("/cars/{id}", updateCar).Methods("PUT")
	router.HandleFunc("/cars/{id}", deleteCar).Methods("DELETE")

	// Start the HTTP server
	log.Fatal(http.ListenAndServe(":8000", router))
}

// Get all cars
func getCars(w http.ResponseWriter, r *http.Request) {
	// Get the MongoDB collection object for cars
	collection := client.Database("rental_cars").Collection("cars")

	// Execute the MongoDB find command to get all cars
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer cur.Close(context.Background())

	// Iterate through the results and add them to an array of cars
	cars := []Car{}
	for cur.Next(context.Background()) {
		var car Car
		err := cur.Decode(&car)
		if err != nil {
			log.Fatal(err)
		}
		cars = append(cars, car)
	}
	// Marshal the slice into JSON and write it to the response.
	jsonBytes, err := json.Marshal(cars)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

// Create a new car
func createCar(w http.ResponseWriter, r *http.Request) {
	// Parse the request body into a Car struct
	var car Car
	err := json.NewDecoder(r.Body).Decode(&car)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Get the MongoDB collection object for cars
	collection := client.Database("rental_cars").Collection("cars")

	// Set the ID of the new car to a new MongoDB ObjectID
	car.ID = primitive.NewObjectID()

	// Set the creation and last update timestamps
	carCreatedAt := time.Now()
	car.UpdatedAt = carCreatedAt

	// Insert the new car into the MongoDB collection
	_, err = collection.InsertOne(context.Background(), car)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Return the new car as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(car)
}

// Get a single car by ID
func getCar(w http.ResponseWriter, r *http.Request) {
	// Get the ID parameter from the URL
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Get the MongoDB collection object for cars
	collection := client.Database("rental_cars").Collection("cars")

	// Execute the MongoDB find command to get the car with the specified ID
	var car Car
	err = collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&car)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	// Return the car as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(car)
}

// Update a single car by ID
func updateCar(w http.ResponseWriter, r *http.Request) {
	// Get the ID parameter from the URL
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Decode the request body into a Car struct
	var carUpdates Car
	err = json.NewDecoder(r.Body).Decode(&carUpdates)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Get the MongoDB collection object for cars
	collection := client.Database("rental_cars").Collection("cars")

	// Get the existing car from the MongoDB collection
	var existingCar Car
	err = collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&existingCar)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	// Update the existing car with the new fields
	existingCar.Model = carUpdates.Model
	// existingCar.Make = carUpdates.Make
	existingCar.Year = carUpdates.Year
	existingCar.Color = carUpdates.Color
	// existingCar.Available = carUpdates.Available
	existingCar.UpdatedAt = time.Now()

	// Execute the MongoDB update command to save the changes
	_, err = collection.ReplaceOne(context.Background(), bson.M{"_id": id}, existingCar)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Return the updated car as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingCar)
}

// Delete a single car by ID
func deleteCar(w http.ResponseWriter, r *http.Request) {
	// Get the ID parameter from the URL
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Get the MongoDB collection object for cars
	collection := client.Database("rental_cars").Collection("cars")

	// Execute the MongoDB delete command to remove the car with the specified ID
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Check if the delete operation actually deleted a car
	if result.DeletedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Car not found"))
		return
	}

	// Return a success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Car deleted"))
}
