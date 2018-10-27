package src

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	f, err := os.Open("./main.go")
	if err != nil {
		panic(err)
	}

	str, _ := ioutil.ReadAll(f)
	fmt.Println(str)
	fmt.Println("vim-go")
}

func dummy() {

	r := mux.NewRouter()

	fmt.Println(r)

}
