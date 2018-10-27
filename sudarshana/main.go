package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type GuruWhatResult struct {
	Srcdir  string `json:"srcdir"`
	Package string `json:"importpath"`
}

type GuruPackageReference_Package struct {
	Path string `json:"path"`
}
type GuruPackageReference struct {
	Desc    string                        `json:"desc"`
	Pos     string                        `json:"pos"`
	Detail  string                        `json:"detail"`
	Package *GuruPackageReference_Package `json:"package"`
}

func main() {
	args := os.Args[1:]

	if len(args) != 2 {
		fmt.Printf("sudarshana [mode=ranks] [file]")
		os.Exit(2)
	}
	mode, file := args[0], args[1]
	// fmt.Printf("mode=%s\n", mode)
	// fmt.Printf("file=%s\n", file)

	switch mode {
	case "ranks":
		output := ranks(file)
		fmt.Printf("%s", output)
	case "popular":
		panic("TODO: Yet to implement")
	case "parse":
		parse(file)
	case "parsefile":
		fileloc := filepath.Base(file)
		dir := filepath.Dir(file)
		parsefile("", dir, fileloc)
	default:
		fmt.Printf("Mode=%s is not recognized", mode)
		os.Exit(2)
	}

}

// This is for sortedCompletitions
// Step 1 - Run guru to find the package information of the source file from godef
// Step 2 - For this package get the sorted list of methods (map of methodName to Count/Priority)
func ranks(inputFile string) string {
	cmd := exec.Command("guru", "-json", "what", inputFile+":#0")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	var result *GuruWhatResult
	json.Unmarshal(out.Bytes(), &result)
	if result.Package == "" {
		return ""
	}
	// TODO: result now has the Package which should be used to identify the sorted list of methods for the package
	output, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("%s", output)
	return string(output)
}

func guru_describe(query string) *GuruPackageReference {
	cmd := exec.Command("guru", "-json", "describe", query)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	// if err != nil {
	// log.Fatal(err)
	// }
	var result *GuruPackageReference
	if err == nil {
		json.Unmarshal(out.Bytes(), &result)
	}
	// if err != nil {
	// log.Fatal(err)
	// }
	// if result.Package == "" {
	// 	return ""
	// }
	// // TODO: result now has the Package which should be used to identify the sorted list of methods for the package
	// output, err := json.Marshal(result)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // fmt.Printf("%s", output)
	// return string(output)
	return result
}
