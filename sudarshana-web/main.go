package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Method struct {
	Name  string
	Count int
}

func main() {

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
	r.Run() // listen and serve on 0.0.0.0:8080
}
