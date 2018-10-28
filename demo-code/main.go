package main

import (
	"log"

	"github.com/ashwanthkumar/devmerge_2k18/demo-code/foo"
)

type Person struct {
	Id   string
	Name string
}

func main() {
	person := Person{"Id", "AP Labs"}
	if !foo.CheckID(person.Name, person.Id) {
		log.Fatalf("Check Id failed")
	}
}
