package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/plandem/xlsx"
)

type Record struct {
	Position    string
	Quantity    int
	Time        float64
	Description string
}

func main() {

	a := app.New()
	w := a.NewWindow("Калькулятор сметы by Peretz489@gmail.com")
	w.Resize(fyne.NewSize(800, 400))
	w.SetFixedSize(true)
	icon, _ := fyne.LoadResourceFromPath("icon.ico")
	w.SetIcon(icon)

	entry := widget.NewMultiLineEntry()
	entry.Resize(fyne.NewSize(780, 300))
	entry.Move(fyne.NewPos(5, 5))
	entry.SetPlaceHolder("Тут будет расчёт")

	var fileData string

	btnFileOpen := widget.NewButton("Выбрать файл", func() {
		dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
			fileData = fmt.Sprint(r.URI())
			fileData = fileData[7:]
			entry.SetText(calculate(r))
			w.SetTitle("Калькулятор сметы by Peretz489@gmail.com: выбран файл " + fileData)
		}, w)
	})
	btnFileOpen.Resize(fyne.NewSize(150, 40))
	btnFileOpen.Move(fyne.NewPos(150, 330))

	btnCalculate := widget.NewButton("Рассчитать", func() {
		if fileData != "" {
			entry.SetText(calculate(fileData))
		}
	})
	btnCalculate.Resize(fyne.NewSize(150, 40))
	btnCalculate.Move(fyne.NewPos(480, 330))

	content := container.NewWithoutLayout(entry, btnFileOpen, btnCalculate)

	w.SetContent(content)

	w.ShowAndRun()

}

func calculate(fileReader interface{}) string {
	if fileReader == nil {
		return "Не указан файл"
	}
	xlsxError:=errors.New("Ошибка открытия файла с прайс-листом или некорректный формат прайса")
	xlsxFile, err := xlsx.Open(fileReader)
	if err != nil {
		return xlsxError.Error()
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
	if len(positions)==0||len(prices)==0||len(quantity)==0||len(time)==0||len(description)==0{
		return xlsxError.Error()
	}
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
	outString := ""
	for idx, position := range positionsInOrder {
		positionHours, positionPrice := positionAttributes(position)
		outString += fmt.Sprintf("%d. %s\n%s\n", idx+1, position.Position, position.Description)
		outString += fmt.Sprintf("%d шт., %0.1f ч., %d руб. с учётом НДС 20%% \n\n", position.Quantity, positionHours, positionPrice)
	}
	remoteOrderPrice, visitOrderPrice := totalCalculation(orderTime)

	orderDays := totalTime(orderTime)

	outString += fmt.Sprintf("ИТОГО:\nУдалённо - %d руб.  с учётом НДС 20%%\nВыезд - %.0f руб. с учётом НДС 20%%. Без учёта трансфера и проживания %0.1f дня.\nПри изменении технического задания, сумма ПНР может быть изменена в большую или меньшую сторону.\n", remoteOrderPrice, visitOrderPrice, orderDays)
	return outString
}

func totalTime(orderTime int) float64 {
	orderDays := math.Round(float64(orderTime)/60./8.*10.) / 10.
	if orderDays < 1 {
		orderDays = 1.
	}
	return orderDays
}

func totalCalculation(orderTime int) (int, float64) {
	remoteOrderPrice := 0
	visitOrderPrice := 0.
	if orderTime < 240 {
		remoteOrderPrice = 240 * 60 * 1.2
	} else {
		remoteOrderPrice = orderTime * 6 * 12
	}

	if orderTime < 480 {
		visitOrderPrice = 480. * 60. * 1.1 * 1.2
	} else {
		visitOrderPrice = float64(remoteOrderPrice) * 1.1
	}
	return remoteOrderPrice, visitOrderPrice
}

func positionAttributes(position Record) (float64, int) {
	positionHours := math.Round(float64(position.Time)/60.*10.) / 10
	positionPrice := int(position.Time * 60. * 1.2)
	return positionHours, positionPrice
}
