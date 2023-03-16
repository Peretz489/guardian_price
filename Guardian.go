package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

var db *DB

type Record struct {
	Position    string
	Price       int
	Description string
}

type DB struct {
	Records map[int]Record `json:"records"`
}

func makeDB() *DB {
	db := DB{
		make(map[int]Record),
	}
	return &db
}

func readDbFile() []byte {
	dbFile, err := os.OpenFile("db.json", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("problems with db.json file: ", err)
	}
	defer dbFile.Close()
	buffer, err := io.ReadAll(dbFile)
	if err != nil {
		log.Fatal("error reading file", err)
	}
	return buffer
}

func getRawData() {
	positionFile, err := os.Open("Positions.txt")
	if err != nil {
		log.Fatal("error opening file", err)
	}
	defer positionFile.Close()
	positions, err := io.ReadAll(positionFile)
	if err != nil {
		log.Fatal("error reading file", err)
	}
	priceFile, err := os.Open("Prices.txt")
	if err != nil {
		log.Fatal("error opening file", err)
	}
	defer priceFile.Close()
	prices, err := io.ReadAll(priceFile)
	if err != nil {
		log.Fatal("error reading file", err)
	}
	descriptionFile, err := os.Open("Description.txt")
	if err != nil {
		log.Fatal("error opening file", err)
	}
	defer positionFile.Close()
	descriptions, err := io.ReadAll(descriptionFile)
	if err != nil {
		log.Fatal("error reading file", err)
	}
	indexFile, err := os.Create("Index.csv")
	if err != nil {
		log.Fatal("cant create index file", err)
	}
	defer indexFile.Close()
	w := csv.NewWriter(indexFile)
	positionData := strings.Split(string(positions), "\n")
	priceData := strings.Split(string(prices), "\n")
	descriptionData := strings.Split(string(descriptions), "\n")
	for idx := range positionData {
		currentPrice := 0
		positionData[idx] = strings.TrimSuffix(positionData[idx], "\r")
		priceData[idx] = strings.TrimSuffix(priceData[idx], "\r")
		descriptionData[idx] = strings.TrimSuffix(descriptionData[idx], "\r")
		if positionData[idx] != "" {
			currentPrice, err = strconv.Atoi(priceData[idx])
			if err != nil {
				log.Fatal(err)
			}
		} else {
			continue
		}
		w.Write([]string{strconv.Itoa(idx + 1), positionData[idx]})
		db.Records[idx+1] = Record{
			positionData[idx],
			currentPrice,
			descriptionData[idx],
		}
	}
	w.Flush()
	saveDB(db)

}

func init() {
	db = makeDB()
	buffer := readDbFile()
	if len(buffer) == 0 {
		getRawData()
	} else {
		err := json.Unmarshal(buffer, db)
		if err != nil {
			log.Fatal("db unmarshall error ", err)
		}
	}
}

func saveDB(db *DB) {
	data, err := json.MarshalIndent(db, "  ", "   ")
	if err != nil {
		log.Println("marshalling error", err)
	}
	dbFile, err := os.OpenFile("db.json", os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("problems with db.json file: ", err)
	}
	defer dbFile.Close()
	_, err = io.WriteString(dbFile, string(data))
	if err != nil {
		log.Fatal("problems with saving to db.json file: ", err)
	}

}

func main() {
	outFile, err := os.Create("out.txt")
	if err != nil {
		log.Fatal("error creating out.txt file", err)
	}
	defer outFile.Close()
	number := 1
	positionNumber, quantity := 0, 0
	outString := ""
	orderTime := 0
	orderPrice := 0
	for {
		fmt.Print("Номер позиции (0 для выхода): ")
		fmt.Scan(&positionNumber)
		if positionNumber == 0 {
			orderDays := math.Round(float64(orderTime)/60./8.*10.) / 10.
			finalString := fmt.Sprintf("ИТОГО:\nУдалённо - %d руб.  с учётом НДС 20%%\nВыезд - %.0f руб. с учётом НДС 20%%. Без учёта трансфера и проживания %0.1f дня.\nПри изменении технического задания, сумма ПНР может быть изменена в большую или меньшую сторону.\n", orderPrice, float64(orderPrice)*1.1, orderDays)
			outFileWrite(outFile, finalString)
			break
		}
		fmt.Print("Количество: ")
		fmt.Scan(&quantity)
		record, ok := db.Records[positionNumber]
		if !ok {
			fmt.Println("Нет такой позиции")
			continue
		}
		workHours, positionPrice := orderPositionAttributes(record, quantity)
		outString = fmt.Sprintf("%d. %s\n%s\n", number, record.Position, record.Description)
		outString += fmt.Sprintf("%d шт., %0.1f ч., %d руб. с учётом НДС 20%% \n\n", quantity, workHours, positionPrice)
		outFileWrite(outFile, outString)

		outString = ""
		number += 1
		orderTime += record.Price * quantity
		orderPrice += positionPrice
	}
}

func outFileWrite(outFile *os.File, outString string) {
	_, err := io.WriteString(outFile, outString)
	if err != nil {
		log.Fatal("Не могу записать данные: ", err)
	}
}

func orderPositionAttributes(record Record, quantity int) (float64, int) {
	workHours := math.Round(float64(record.Price*quantity)/60.*10.) / 10
	positionPrice := int(float64(record.Price*quantity*60) * 1.2)
	return workHours, positionPrice
}
