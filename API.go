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

	_ "github.com/mattn/go-sqlite3"
)

type exchange_rates struct {
	id         int
	currency   string
	rate       float32
	created_at string
}
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

	//ticker := time.NewTicker(time.Minute * 1)

	/*for {
	log.Println("Start iteration")
	if time.Now().Hour() != 0 && time.Now().Minute() != 0 {
		time.Sleep(1 * time.Second)
		continue
	}*/
	rate, err := getExchangeRate(apiKey)
	if err != nil {
		log.Println(err)
		//continue
	}
	fmt.Println(rate)

	Curr, err := saveExchangeRate(rate)
	if err != nil {
		log.Println(err)
		//continue
	}
	log.Printf("Exchange rate saved: %f", rate)
	for _, Cur := range Curr {
		fmt.Println(Cur.id, Cur.currency, Cur.rate, Cur.created_at)
	}
	//}
}

func getExchangeRate(apiKey string) (float64, error) {

	type Rates struct {
		RUB float64 `json:"RUB"`
	}
	type Response struct {
		Rates Rates `json:"rates"`
	}

	apiURL := "https://api.apilayer.com/fixer/latest?base=USD&symbols=RUB"
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

func saveExchangeRate(rate float64) ([]exchange_rates, error) {
	db, err := sql.Open("sqlite3", "Current.db")
	if err != nil {
		panic(err)
	}
	fmt.Println("Подключение с БД установленно!")
	defer db.Close()
	stmt, err := db.Prepare("insert into exchange_rates (currency, rate, created_at) values (?, ?, ?)")
	if err != nil {
		panic(err)
	}
	_, err = stmt.Exec("USD", rate, time.Now().Format("2006-01-02"))
	if err != nil {
		panic(err)
	}
	rows, err := db.Query("select * from exchange_rates")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	currenct := []exchange_rates{}

	for rows.Next() {
		Cur := exchange_rates{}
		err := rows.Scan(&Cur.id, &Cur.currency, &Cur.rate, &Cur.created_at)
		if err != nil {
			panic(err)
			continue
		}
		currenct = append(currenct, Cur)
	}
	return currenct, nil
}
