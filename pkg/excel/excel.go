package excel

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"inspense-bot/database"
	"strconv"
	"strings"
	"time"
)

const (
	rowOffset      = 9
	columnOffset   = 'E'
	blackColor     = "000000"
	incomeColor    = "1B79AD"
	expenseColor   = "1D7B7D"
	totalColor     = "189855"
	monthFillColor = "1B79AD"
	monthFontColor = "FFFFFF"
)

func DetailedReport(finances map[string][]database.Finance, ft string, im ...bool) ([]byte, error) {
	isMonth := len(im) > 0 && im[0]

	file, err := excelize.OpenFile("./assets/template.xlsx")
	if err != nil {
		return nil, err
	}

	if ft == "income" {
		err = file.DeleteSheet("expense")
	} else {
		err = file.DeleteSheet("income")
	}
	if err != nil {
		return nil, err
	}

	if isMonth {
		file.SetCellInt(ft, "Q6", int(time.Now().Month()))

		monthStyles, err := file.NewStyle(&excelize.Style{
			Fill: excelize.Fill{
				Type:  "pattern",
				Color: []string{monthFillColor}, Pattern: 1},
			Border:    []excelize.Border{{Type: "bottom", Style: 2, Color: blackColor}},
			Font:      &excelize.Font{Family: "Gill Sans MT", Size: 12, Bold: true, Color: monthFontColor},
			Alignment: &excelize.Alignment{Horizontal: "center"},
		})
		if err != nil {
			return nil, err
		}
		file.SetCellStyle(ft, "Q6", "Q6", monthStyles)
	}

	file.SetCellInt(ft, "R6", time.Now().Year())

	borderFill := excelize.Fill{Type: "pattern", Color: []string{incomeColor}, Pattern: 1}
	if ft == "expense" {
		borderFill.Color = []string{expenseColor}
	}
	borderStyles, err := file.NewStyle(&excelize.Style{Fill: borderFill})
	if err != nil {
		return nil, err
	}

	dateBorder := []excelize.Border{
		{Type: "top", Style: 2, Color: blackColor},
		{Type: "bottom", Style: 2, Color: blackColor},
		{Type: "left", Style: 2, Color: blackColor},
		{Type: "right", Style: 2, Color: blackColor},
	}

	categoryStyles, err := file.NewStyle(&excelize.Style{
		Border: dateBorder,
		Font: &excelize.Font{
			Bold:   true,
			Family: "Calisto MT",
			Size:   12,
		},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	if err != nil {
		return nil, err
	}

	dateStyles, err := file.NewStyle(&excelize.Style{
		Border: dateBorder,
		Font:   &excelize.Font{Color: blackColor},
	})
	if err != nil {
		return nil, err
	}

	var (
		monthStep  = columnOffset
		categories = make(map[string]map[int32]float64)
	)

	for month, fin := range finances {
		monthCode, _ := strconv.Atoi(month)
		cellPlacement := fmt.Sprintf("%s8", string(monthStep))

		if isMonth {
			monthCode, _ = strconv.Atoi(month)
			file.SetCellValue(ft, cellPlacement, monthCode)
		} else {
			file.SetCellValue(ft, cellPlacement, strings.ToUpper(time.Month(monthCode).String()[:3]))
		}

		file.SetCellStyle(ft, cellPlacement, cellPlacement, dateStyles)

		for _, finance := range fin {
			category := finance.Category
			amount := finance.Amount

			if _, ok := categories[category]; !ok {
				categories[category] = make(map[int32]float64)
			}

			categories[category][monthStep] += amount
		}
		monthStep++
	}

	file.SetCellValue(ft, fmt.Sprintf("%s8", string(monthStep)), "TOTAL")
	file.SetCellStyle(ft, fmt.Sprintf("%s8", string(monthStep)), fmt.Sprintf("%s8", string(monthStep)), dateStyles)

	amountFont := &excelize.Font{Family: "Cascadia Mono", Size: 11}

	amountStyles, err := file.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: blackColor},
			{Type: "bottom", Style: 1, Color: blackColor},
			{Type: "left", Style: 2, Color: blackColor},
			{Type: "right", Style: 2, Color: blackColor},
		},
		Font:      amountFont,
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})
	if err != nil {
		return nil, err
	}

	offset := rowOffset
	for category, monthlyAmounts := range categories {
		file.SetCellStyle(ft, fmt.Sprintf("C%d", offset), fmt.Sprintf("C%d", offset), borderStyles)
		file.SetCellStyle(ft, fmt.Sprintf("D%d", offset), fmt.Sprintf("D%d", offset), categoryStyles)
		file.SetCellValue(ft, fmt.Sprintf("D%d", offset), category)

		for month, amount := range monthlyAmounts {
			file.SetCellValue(ft, fmt.Sprintf("%c%d", month, offset), amount)
			file.SetCellStyle(ft, fmt.Sprintf("%s%d", string(monthStep), offset), fmt.Sprintf("%s%d", string(monthStep), offset), amountStyles)
		}
		offset++
	}

	monthStep = columnOffset
	for i := 0; i < len(finances); i++ {
		for j := rowOffset; j < len(categories)+rowOffset; j++ {
			file.SetCellStyle(ft, fmt.Sprintf("%s%d", string(monthStep), j), fmt.Sprintf("%s%d", string(monthStep), j), amountStyles)
		}
		monthStep++
	}

	amountFont.Color = totalColor
	bottomStyles, err := file.NewStyle(&excelize.Style{
		Border:    dateBorder,
		Font:      amountFont,
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})
	if err != nil {
		return nil, err
	}

	file.SetCellValue(ft, fmt.Sprintf("D%d", offset), "Total")
	file.SetCellStyle(ft, fmt.Sprintf("D%d", offset), fmt.Sprintf("D%d", offset), categoryStyles)

	monthStep = columnOffset
	for i := 0; i < len(finances)+1; i++ {
		file.SetCellStyle(ft, fmt.Sprintf("%s%d", string(monthStep), offset), fmt.Sprintf("%s%d", string(monthStep), offset), bottomStyles)
		file.SetCellFormula(ft, fmt.Sprintf("%s%d", string(monthStep), offset), fmt.Sprintf("SUM(%s9:%s%d)", string(monthStep), string(monthStep), offset-1))
		monthStep++
	}

	var (
		series   []excelize.ChartSeries
		monthEnd = columnOffset + len(finances) - 1
	)

	for j := rowOffset; j < len(categories)+rowOffset; j++ {
		series = append(series, excelize.ChartSeries{
			Name:       fmt.Sprintf("%s!$D$%d", ft, j),
			Categories: fmt.Sprintf("%s!$E$8:$%s$8", ft, string(monthEnd)),
			Values:     fmt.Sprintf("%s!$E$%d:$%s$%d", ft, j, string(monthEnd), j),
		})

		file.SetCellFormula(ft, fmt.Sprintf("%s%d", string(monthEnd+1), j), fmt.Sprintf("SUM(E%d:%s%d)", j, string(monthEnd), j))
		file.SetCellStyle(ft, fmt.Sprintf("%s%d", string(monthEnd+1), j), fmt.Sprintf("%s%d", string(monthEnd+1), j), bottomStyles)
	}

	period := "Yearly"
	if isMonth {
		period = "Monthly"
	}

	if err := file.AddChart(ft, fmt.Sprintf("%c8", monthEnd+3), &excelize.Chart{
		Type: excelize.Col3DClustered, Series: series,
		Title:     excelize.ChartTitle{Name: fmt.Sprintf("%s report of %ss", period, strings.Title(ft))},
		Dimension: excelize.ChartDimension{Width: 620, Height: 400},
		XAxis:     excelize.ChartAxis{Font: excelize.Font{Color: blackColor}},
		YAxis:     excelize.ChartAxis{Font: excelize.Font{Color: blackColor}},
	}); err != nil {
		return nil, err
	}

	buff, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	if err := file.Close(); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
