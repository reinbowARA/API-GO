package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	apikey := "Xa0PYkfM35Sz02AFPkjPRXCIDKb2MXJj"

	var today string
	today = time.Now().Format("2006-01-02")
	fmt.Println(today)

	apiURL := fmt.Sprintf("https://api.apilayer.com/fixer/%s?base=USD&symbols=RUB", today)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("apikey", apikey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	fmt.Println(string(body))
}
