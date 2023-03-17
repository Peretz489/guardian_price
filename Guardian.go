package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/plandem/xlsx"
)

type Record struct {
	Position    string
	Quantity    int
	Time        float64
	Description string
}

func main() {
	//xlsxFile := xlsx.New()
	xlsxFile, err := xlsx.Open("data.xlsx")
	if err != nil {
		log.Fatal("Cant open *.xlsx file ", err)
	}
	defer xlsxFile.Close()
	sheet := xlsxFile.Sheet(0)
	positions := sheet.Col(0).Values()
	prices := sheet.Col(2).Values()
	quantity := sheet.Col(3).Values()
	time := sheet.Col(4).Values()
	description := sheet.Col(5).Values()
	positionsInOrder := make([]Record, 0)
	orderTime := 0
	for idx := range positions {
		positionTime, err := strconv.Atoi(time[idx])
		if err != nil || positionTime == 0 || prices[idx] == "" {
			continue
		}
		if quantity[idx] != "" {
			positionQuantity, err := strconv.Atoi(quantity[idx])
			if err != nil {
				log.Fatal(err)
			}
			positionsInOrder = append(positionsInOrder, Record{positions[idx], positionQuantity, float64(positionTime), description[idx]})
			orderTime += positionTime
		}
	}
	outFile, err := os.Create("out.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	outString := ""
	orderPrice := 0
	for idx, position := range positionsInOrder {
		positionHours, positionPrice := positionAttributes(position)
		orderPrice += positionPrice
		outString += fmt.Sprintf("%d. %s\n%s\n", idx+1, position.Position, position.Description)
		outString += fmt.Sprintf("%d шт., %0.1f ч., %d руб. с учётом НДС 20%% \n\n", position.Quantity, positionHours, positionPrice)
	}
	remoteOrderPrice := 0
	visitOrderPrice := 0.
	if orderTime < 240 {
		remoteOrderPrice = 240 * 60 * 1.2
	} else {
		remoteOrderPrice = orderPrice
	}

	if orderTime < 480 {
		visitOrderPrice = 480. * 60. * 1.1 * 1.2
	} else {
		visitOrderPrice = float64(remoteOrderPrice) * 1.1
	}

	orderDays := math.Round(float64(orderTime)/60./8.*10.) / 10.

	outString += fmt.Sprintf("ИТОГО:\nУдалённо - %d руб.  с учётом НДС 20%%\nВыезд - %.0f руб. с учётом НДС 20%%. Без учёта трансфера и проживания %0.1f дня.\nПри изменении технического задания, сумма ПНР может быть изменена в большую или меньшую сторону.\n", remoteOrderPrice, visitOrderPrice, orderDays)
	fmt.Println(outString)
	outFileWrite(outFile, outString)
}

func outFileWrite(outFile *os.File, outString string) {
	_, err := io.WriteString(outFile, outString)
	if err != nil {
		log.Fatal("Не могу записать данные: ", err)
	}
}

func positionAttributes(position Record) (float64, int) {
	positionHours := math.Round(float64(position.Time)/60.*10.) / 10
	positionPrice := int(position.Time * 60. * 1.2)
	return positionHours, positionPrice
}
