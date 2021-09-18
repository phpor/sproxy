package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	resp, err := http.Get(os.Args[1])
	if err != nil {
		panic(err)
	}
	res, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}
	fmt.Println(string(res))

}
