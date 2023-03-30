package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type exchange_rates struct {
	id         int
	currency   string
	rate       float64
	created_at time.Time
}
type APIkey struct {
	Key    string `json:"key"`
	Keyimg string `json:"keyimg"`
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
	apikeyimg := key.Keyimg

	var TimeWait int

	if time.Now().Minute() > 30 {
		TimeWait = int(60.0 - float64(time.Now().Minute()))
	} else {
		TimeWait = int(30.0 - float64(time.Now().Minute()))
	}
	fmt.Println("Start program!")
	fmt.Printf("Please wait %d min \n", TimeWait)

	for {
		if time.Now().Minute()%10 == 0 && time.Now().Second() == 0 {
			log.Println("Start read API")
		rep:
			rate, err := getExchangeRate(apiKey)
			if err != nil {
				log.Println(err)
				goto rep
			}
			log.Println("API read successfully")
			Curr, err := saveExchangeRate(rate)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Printf("Exchange rate saved: %f", rate)
			GetPngGraph(Curr)
			UploadImg(apikeyimg)
			log.Printf("stand %d minutes \n", TimeWait)
			time.Sleep(5 * time.Minute) // кастыль, инача 2 раза подряд считывает цикл
		}

	}
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
	log.Println("Подключение с БД установленно!")
	defer db.Close()
	stmt, err := db.Prepare("insert into exchange_rates (currency, rate, created_at) values (?, ?, ?)")
	if err != nil {
		panic(err)
	}
	_, err = stmt.Exec("USD", rate, time.Now().Format("2006-01-02 15:04:05"))
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

	point, err := plotter.NewScatter(pts)
	if err != nil {
		panic(err)
	}
	// устанавливаем точки на красный цвет
	point.GlyphStyle.Color = plotutil.Color(0)

	p.Add(point)
	// Устанавливаем метки на оси X
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02 15:04:05"}

	// Сохранение графика в файл
	if err := p.Save(10*vg.Inch, 10*vg.Inch, "src/usd_to_rub.png"); err != nil {
		panic(err)
	}
}

func UploadImg(APIkey string) {
	// открываем файл с изображением
	file, err := os.Open("src/usd_to_rub.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// создаем новый буфер для формы
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// добавляем файл в форму
	part, err := writer.CreateFormFile("image", "usd_to_rub.png")
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		panic(err)
	}
	// закрываем форму
	writer.Close()

	// создаем новый POST-запрос на API imgbb.com
	url := fmt.Sprintf("https://api.imgbb.com/1/upload?&key=%s", APIkey)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// отправляем запрос и получаем ответ
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

}
