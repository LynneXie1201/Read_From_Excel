package main

import (
	"fmt"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type person struct {
	Name  string
	Phone string
}

func main() {

	session, err := mgo.Dial("localhost") // open an connection -> Dial function
	if err != nil {                       //  if you have a
		panic(err)
	}
	defer session.Close() // session must close at the end

	session.SetMode(mgo.Monotonic, true) // Optional. Switch the session to a monotonic behavior.

	c := session.DB("mgo1").C("contacts")
	err = c.Insert(&person{"Ale", "+55 53 8116 9639"},
		&person{"Cla", "+55 53 8402 8510"}, &person{"Lynne", "647 879 6642"})
	if err != nil {
		log.Fatal(err)
	}

	result := person{}
	err = c.Find(bson.M{"name": "Ale"}).One(&result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Phone:", result.Phone)
	//fdsafsafsa

}
