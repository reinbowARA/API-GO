package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type CurrencyListResponse struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
}

type CurrencyRange struct {
	Status  int               `json:"status"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data"`
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

	var pairs [92]string

	//получаем количество пар
	for value, pair := range currencyListResponse.Data {
		fmt.Println(value, " = ", string(pair))
		pairs[value] = string(pair)
	}

	currentDUO := strings.Join(pairs[:], ",")

	url = fmt.Sprintf("https://currate.ru/api/?get=rates&pairs=%s&key=%s", currentDUO, apiKey)

	resp, err = http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var currencyRange CurrencyRange

	err = json.Unmarshal(body, &currencyRange)
	if err != nil {
		panic(err)
	}

	if currencyRange.Status != 200 {
		panic(currencyListResponse.Message)
	}

	//получаем количество пар с расценкой на сегодняшний день (по умолчанию берутся последние данные (GMT +03:00))
	for pair, rate := range currencyRange.Data {
		fmt.Println(string(pair), "=", string(rate))
	}
}
