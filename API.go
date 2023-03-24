package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	apikey := "Xa0PYkfM35Sz02AFPkjPRXCIDKb2MXJj"
	apiURL := "https://api.apilayer.com/fixer/timeseries?base=USD&symbols=RUB&start_date=2023-03-01&end_date=2023-03-05"

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
