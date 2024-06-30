package main

import (
	"fmt"
	"log"
)

func main() {
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}
	// print connection
	// fmt.Printf("%+v\n", store)

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewAPIServer(":8084", store)
	server.Run()
	fmt.Println("Hereee")
}
