package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

type GuruResult struct {
	Srcdir     string `json:"srcdir"`
	ImportPath string `json:"importpath"`
}

func main() {
	// Step 1 - Run guru to find the package information of the source file from godef
	// Step 2 - For this package get the sorted list of methods (map of methodName to Count/Priority)

	// This is for sortedCompletitions
	cmd := exec.Command("guru", "-json", "what", "/Users/ashwanthkumar/hacks/go-workspace/src/github.com/ashwanthkumar/devmerge_2k18/sudarshana/main.go"+":#0")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	var result *GuruResult
	// fmt.Printf("%q\n", out.String())
	json.Unmarshal(out.Bytes(), &result)
	fmt.Println(result)
}
