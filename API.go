package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CurrencyListResponse struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
}

func main() {
	apiKey := "49293d12f95c00b54dab08a4274390d7"
	url := fmt.Sprintf("https://currate.ru/api/?get=currency_list&key=%s", apiKey)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var currencyListResponse CurrencyListResponse
	err = json.Unmarshal(body, &currencyListResponse)
	if err != nil {
		panic(err)
	}

	if currencyListResponse.Status != "200" {
		panic(currencyListResponse.Message)
	}

	for value, pair := range currencyListResponse.Data {
		fmt.Println(value+1, " = ", string(pair))
	}
}
