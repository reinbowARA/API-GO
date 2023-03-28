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
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type exchange_rates struct {
	id         int
	currency   string
	rate       float64
	created_at time.Time
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
	GetPngGraph(Curr)
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
			//continue
		}
		currenct = append(currenct, Cur)
	}
	return currenct, nil
}
func GetPngGraph(currents []exchange_rates) {
	// Создание нового графика
	p := plot.New()

	// Установка заголовка графика
	p.Title.Text = "USD to RUB Exchange Rate"

	// Установка меток на осях координат
	p.X.Label.Text = "Date"
	p.Y.Label.Text = "Exchange Rate"

	var CreatedAt []time.Time
	for _, Currate := range currents {
		Cur := Currate.created_at
		CreatedAt = append(CreatedAt, Cur)
	}

	/*var times []time.Time
	for _, Cur := range CreatedAt {
		t, _ := time.Parse("2006-01-02", Cur)
		times = append(times, t)
	}*/

	var rate []float64
	for _, Currate := range currents {
		Cur := Currate.rate
		rate = append(rate, Cur)
	}

	// Создание точек для графика
	pts := make(plotter.XYs, len(rate), len(CreatedAt))
	for i := range pts {
		pts[i].Y = rate[i]

	}

	for i := range pts {
		pts[i].X = float64(CreatedAt[i].Unix())
	}
	// Создание линии для графика
	line, err := plotter.NewLine(pts)
	if err != nil {
		panic(err)
	}

	// Добавление линии на график
	p.Add(line)

	// Устанавливаем метки на оси X
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02"}

	// Сохранение графика в файл
	if err := p.Save(20*vg.Inch, 20*vg.Inch, "src/usd_to_rub.png"); err != nil {
		panic(err)
	}
}
