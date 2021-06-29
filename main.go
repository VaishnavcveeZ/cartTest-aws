package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"
	"io/ioutil"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type product struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Pname       string             `json:"Pname,omitempty" bson:"Pname,omitempty"`
	Price       string             `json:"Price,omitempty" bson:"Price,omitempty"`
	Description string             `json:"Description,omitempty" bson:"Description,omitempty"`
	Image       string             `json:"Image,omitempty" bson:"Image,omitempty"`
}

//globals
var templates *template.Template
var client *mongo.Client

func main() {
	fmt.Println("Server running... ")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	templates = template.Must(template.ParseGlob("templates/*.html"))
	http.FileServer(http.Dir("./static/"))

	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler((http.StripPrefix("/static",  http.FileServer(http.Dir("."+"/static")))))

	router.HandleFunc("/", index).Methods("GET")
	router.HandleFunc("/additem", additem).Methods("GET")
	router.HandleFunc("/users", users).Methods("GET")
	router.HandleFunc("/add", add).Methods("POST")
	router.HandleFunc("/cartin/{pname}", cartin).Methods("GET")
	router.HandleFunc("/cart", cart).Methods("GET")

	http.Handle("/", router)
	http.ListenAndServe(":4000", nil)
}

func index(response http.ResponseWriter, request *http.Request) {

	collection := client.Database("CartUser").Collection("product")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cur, _ := collection.Find(ctx, bson.M{})
	var result bson.D
	
	var out =[]map[string]string{}
	for cur.Next(ctx) {
		data_map :=map[string]string{}
		cur.Decode(&result)
		for i:=1;i<5;i++{
			key :=result[i].Key
			value := fmt.Sprintf("%v", result[i].Value)
			data_map[key]=value
		}
		out = append(out, data_map)
	}

	templates.ExecuteTemplate(response, "index.html", out)
}
func additem(response http.ResponseWriter, request *http.Request) {
	templates.ExecuteTemplate(response, "additem.html", nil)
}

func add(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("contet-type", "application/json")
	var product product
	product.Pname = request.FormValue("pname")
	product.Price = request.FormValue("price")
	product.Description = request.FormValue("descr")
	file, _, _ := request.FormFile("pimage")
	var buf bytes.Buffer
	io.Copy(&buf, file)
	// content := buf.String()
	// fmt.Printf("file size : %v \n", content)
	writeName := "static/product-images/" + product.Pname + ".jpg"
	ioutil.WriteFile(writeName, buf.Bytes(), 0600)
	//product.Image = writeName

	fmt.Println(product)
	json.NewDecoder(request.Body).Decode(&product)
	collection := client.Database("CartUser").Collection("product")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection.InsertOne(ctx, product)
	http.Redirect(response, request, "/", http.StatusSeeOther)
}

func users(response http.ResponseWriter, request *http.Request) {
	collection := client.Database("CartUser").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cur, _ := collection.Find(ctx, bson.M{})
	var result bson.D
	data := []string{}

	for cur.Next(ctx) {
		cur.Decode(&result)
		str := fmt.Sprintf("%v", result[1].Value)
		data = append(data, str)
		//fmt.Println(result)
	}

	templates.ExecuteTemplate(response, "users.html", data)
}
func cartin(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("contet-type", "application/json")
	params := mux.Vars(request)
	pname:=params["pname"]
	var Product product
	collection := client.Database("CartUser").Collection("product")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection.FindOne(ctx, product{Pname: pname}).Decode(&Product)
	collection2 := client.Database("TheCart").Collection("user")
	collection2.InsertOne(ctx, Product)

	http.Redirect(response, request, "/", http.StatusSeeOther)
}
func cart(response http.ResponseWriter, request *http.Request) {

	collection := client.Database("TheCart").Collection("user")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cur, _ := collection.Find(ctx, bson.M{})
	var result bson.D
	
	var out =[]map[string]string{}
	for cur.Next(ctx) {
		data_map :=map[string]string{}
		cur.Decode(&result)
		for i:=1;i<5;i++{
			key :=result[i].Key
			value := fmt.Sprintf("%v", result[i].Value)
			data_map[key]=value
		}
		out = append(out, data_map)
	}
	templates.ExecuteTemplate(response, "cart.html", out)
}