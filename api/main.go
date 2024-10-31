package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Movie struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Title    string             `bson:"title"`
	Year     int                `bson:"year"`
	Director string             `bson:"director"`
	Created  time.Time          `bson:"created"`
	Updated  time.Time          `bson:"updated"`
}

var client *mongo.Client
var collection *mongo.Collection

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MongoDB URI not defined!")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB: " + mongoURI)

	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	outputMsg := "Available databases ("
	outputMsg += strconv.Itoa(len(databases)) + "): "
	for _, db := range list[:len(databases)-1] {
		outputMsg += db + ", "
	}
	outputMsg += list[len(databases)-1]
	log.Print(outputMsg)

	//	log.Println(databases)

	collection = client.Database("Movies").Collection("moviesList")

	// Setup router
	r := mux.NewRouter()
	r.HandleFunc("/movies", createMovie).Methods("POST")
	r.HandleFunc("/movies", getMovie).Methods("GET")
	r.HandleFunc("/movies/{id}", getMovie).Methods("GET")
	r.HandleFunc("/movies/{id}", updateMovie).Methods("PUT")
	r.HandleFunc("/movies/{id}", deleteMovie).Methods("DELETE")

	log.Printf("API Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func createMovie(w http.ResponseWriter, r *http.Request) {
	var movie Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	movie.Created = time.Now()
	movie.Updated = time.Now()

	result, err := collection.InsertOne(context.Background(), movie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	movie.ID = result.InsertedID.(primitive.ObjectID)
	json.NewEncoder(w).Encode(movie)
	log.Println("createMovie " + r.Method + " " + r.URL.Path)
}

func getMovie(w http.ResponseWriter, r *http.Request) {
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var movies []Movie
	if err = cursor.All(context.Background(), &movies); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(movies)
	log.Println("getMovie " + r.Method + " " + r.URL.Path)
}

func updateMovie(w http.ResponseWriter, r *http.Request) {
	id, _ := primitive.ObjectIDFromHex(mux.Vars(r)["id"])
	var movie Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	movie.Updated = time.Now()
	//	update := bson.M{"$set": bson.M{"data": movie.Data, "updated": movie.Updated}}
	update := bson.M{"$set": bson.M{"title": movie.Title, "year": movie.Year, "director": movie.Director, "updated": movie.Updated}}

	result := collection.FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": id},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	if err := result.Decode(&movie); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(movie)
	log.Println("updateMovie " + r.Method + " " + r.URL.Path)
}

func deleteMovie(w http.ResponseWriter, r *http.Request) {
	id, _ := primitive.ObjectIDFromHex(mux.Vars(r)["id"])
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Println("deleteMovie " + r.Method + " " + r.URL.Path)
}
