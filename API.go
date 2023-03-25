package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type APIkey struct {
	Key string `json:"key"`
}

func main() {
	file, err := os.Open("apikey.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var key APIkey
	Key := json.NewDecoder(file)
	err = Key.Decode(&key)
	if err != nil {
		panic(err)
	}

	apiKey := key.Key // получаем api из json файла {"key":"token_API"}

	//ticker := time.NewTicker(time.Minute * 10)

	//for range ticker.C {
	rate, err := getExchangeRate(apiKey)
	if err != nil {
		log.Println(err)
		//continue
	}
	fmt.Println(rate)
	err = saveExchangeRate(rate)
	if err != nil {
		log.Println(err)
		//continue
	}
	log.Printf("Exchange rate saved: %f", rate)
	//}
}

func getExchangeRate(apiKey string) (float64, error) {

	type Rates struct {
		RUB float64 `json:"RUB"`
	}
	type Response struct {
		Rates Rates `json:"rates"`
	}

	apiURL := fmt.Sprintf("https://api.apilayer.com/fixer/latest?base=USD&symbols=RUB")
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("apikey", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var date Response
	err = json.Unmarshal(body, &date)
	return date.Rates.RUB, err
}

func saveExchangeRate(rate float64) error {
	db, err := sql.Open("sqlite3", "Current.db")
	if err != nil {
		return err
	}
	fmt.Println("Подключение с БД установленно!")
	defer db.Close()
	stmt, err := db.Prepare("insert into Current.exchange_rates (currency, rate, created_at) values ($1, $2, $3)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec("USD", rate, time.Now())
	if err != nil {
		return err
	}
	return nil
}
