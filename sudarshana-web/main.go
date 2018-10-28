package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Method struct {
	Name  string
	Count int
}

type MethodSample struct {
	Name string
	Code string
}

type GuruWhatResult struct {
	Srcdir  string `json:"srcdir"`
	Package string `json:"importpath"`
}

func readAndPopulatePopularPatterns() map[string][]MethodSample {
	popularPatterns := make(map[string][]MethodSample)
	popularData := "popular_patterns.tsv"

	file, err := os.Open(popularData)
	if err != nil {
		log.Fatalf(err.Error())
	}
	csvReader := csv.NewReader(file)
	csvReader.Comma = '\t'
	csvReader.LazyQuotes = true
	for {
		fields, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		packageName := fields[0]
		funcName := fields[1]
		funcCode := strings.Replace(fields[2], "\\", "", -1)
		method := MethodSample{
			Name: funcName,
			Code: funcCode,
		}

		key := fmt.Sprintf("%s#%s", packageName, funcName)
		existing, _ := popularPatterns[key]
		popularPatterns[key] = append(existing, method)
	}

	return popularPatterns
}

func readAndPopulateRankedCompletions() map[string][]Method {
	rankedCompletions := make(map[string][]Method)
	rankedData := "ranked-completions.tsv"
	file, err := os.Open(rankedData)
	if err != nil {
		log.Fatalf(err.Error())
	}
	csvReader := csv.NewReader(file)
	csvReader.Comma = '\t'
	for {
		fields, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		packageName := fields[0]
		funcName := fields[1]
		funcCount, _ := strconv.Atoi(fields[2])
		method := Method{
			Name:  funcName,
			Count: funcCount,
		}

		existing, present := rankedCompletions[packageName]
		if present {
			rankedCompletions[packageName] = append(existing, method)
		} else {
			rankedCompletions[packageName] = append(existing, method)
		}
	}

	return rankedCompletions
}

func main() {

	rankedCompletions := readAndPopulateRankedCompletions()
	popularPatterns := readAndPopulatePopularPatterns()

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	r.GET("/ranked", func(c *gin.Context) {
		inputPackage := c.Query("package")
		methods, present := rankedCompletions[inputPackage]
		output := make([]Method, 0)
		if present {
			output = methods
		}
		c.JSON(200, gin.H{
			"result": output,
		})
	})
	r.GET("/popular", func(c *gin.Context) {
		inputPackage := c.Query("package")
		inputFunc := c.Query("func")
		key := fmt.Sprintf("%s#%s", inputPackage, inputFunc)
		methods, present := popularPatterns[key]
		output := make([]MethodSample, 0)
		if present {
			output = methods
		}
		c.JSON(200, gin.H{
			"result": output,
		})
	})
	r.GET("/popularForVSCode", func(c *gin.Context) {
		inputFile := c.Query("file")
		inputFunc := c.Query("func")
		result := packageOf(inputFile)

		packageToUse := result.Package
		fmt.Printf("%s\n", packageToUse)
		indexOfFunc := strings.Index(inputFunc, " func")
		funcToUse := inputFunc[0:indexOfFunc]
		fmt.Printf("FuncToUse=%s\n", funcToUse)

		key := fmt.Sprintf("%s#%s", packageToUse, funcToUse)
		methods, present := popularPatterns[key]
		output := make([]MethodSample, 0)
		if present {
			output = methods
		}
		c.JSON(200, gin.H{
			"result": output,
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}

func packageOf(inputFile string) *GuruWhatResult {
	fileParts := strings.Split(inputFile, ":")
	cmd := exec.Command("guru", "-json", "what", fileParts[0]+":#0")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	var result *GuruWhatResult
	json.Unmarshal(out.Bytes(), &result)
	if result.Package == "" {
		return &GuruWhatResult{}
	}
	// TODO: result now has the Package which should be used to identify the sorted list of methods for the package
	// output, err := json.Marshal(result)
	// if err != nil {
	// log.Fatal(err)
	// }
	// fmt.Printf("%s", output)
	// return string(output)
	return result
}
