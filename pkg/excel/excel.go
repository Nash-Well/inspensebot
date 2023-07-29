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
	rowOffset    = 9
	columnOffset = 'E'
)

// TODO: Optimize styles
// TODO: Remove meaningless variables make code much readable
// TODO: Break to smaller pieces

func YearlyReport(finances map[string][]database.Finance, ft string) ([]byte, error) {
	file, err := excelize.OpenFile("./assets/yearly.xlsx")
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

	borderStyles, err := file.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"1B79AD"}, Pattern: 1},
	})
	if ft == "expense" {
		borderStyles, err = file.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{"1D7B7D"}, Pattern: 1},
		})
	}
	if err != nil {
		return nil, err
	}

	categoryStyles, err := file.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "bottom", Style: 2, Color: "000000"},
			{Type: "top", Style: 2, Color: "000000"},
			{Type: "left", Style: 2, Color: "000000"},
		},
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

	monthStyles, err := file.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "top", Style: 2, Color: "000000"},
			{Type: "bottom", Style: 2, Color: "000000"},
			{Type: "left", Style: 2, Color: "000000"},
			{Type: "right", Style: 2, Color: "000000"},
		},
		Font: &excelize.Font{Color: "000000"},
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
		file.SetCellValue(ft, fmt.Sprintf("%s8", string(monthStep)), strings.ToUpper(time.Month(monthCode).String()[:3]))
		file.SetCellStyle(ft, fmt.Sprintf("%s8", string(monthStep)), fmt.Sprintf("%s8", string(monthStep)), monthStyles)

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

	file.SetCellValue(ft, fmt.Sprintf("%s8", string(monthStep)), "YEAR")
	file.SetCellStyle(ft, fmt.Sprintf("%s8", string(monthStep)), fmt.Sprintf("%s8", string(monthStep)), monthStyles)

	amountStyles, err := file.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: "000000"},
			{Type: "bottom", Style: 1, Color: "000000"},
			{Type: "left", Style: 2, Color: "000000"},
			{Type: "right", Style: 2, Color: "000000"},
		},
		Font: &excelize.Font{
			Family: "Cascadia Mono",
			Size:   11,
		},
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})
	if err != nil {
		return nil, err
	}

	offset := rowOffset
	for category, monthlyAmounts := range categories {
		file.SetCellStyle(ft, fmt.Sprintf("C%v", offset), fmt.Sprintf("C%v", offset), borderStyles)
		file.SetCellStyle(ft, fmt.Sprintf("D%v", offset), fmt.Sprintf("D%v", offset), categoryStyles)
		file.SetCellValue(ft, fmt.Sprintf("D%v", offset), category)

		for month, amount := range monthlyAmounts {
			file.SetCellValue(ft, fmt.Sprintf("%c%d", month, offset), amount)
			file.SetCellStyle(ft, fmt.Sprintf("%s%v", string(monthStep), offset), fmt.Sprintf("%s%v", string(monthStep), offset), amountStyles)
		}
		offset++
	}

	// Set styles for amount
	monthStep = columnOffset
	for i := 0; i < len(finances); i++ {
		for j := rowOffset; j < len(categories)+rowOffset; j++ {
			file.SetCellStyle(ft, fmt.Sprintf("%v%v", string(monthStep), j), fmt.Sprintf("%v%v", string(monthStep), j), amountStyles)
		}
		monthStep++
	}

	// Reset last row styles
	bottomStyles, err := file.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "top", Style: 2, Color: "000000"},
			{Type: "bottom", Style: 2, Color: "000000"},
			{Type: "left", Style: 2, Color: "000000"},
			{Type: "right", Style: 2, Color: "000000"},
		},
		Font: &excelize.Font{
			Family: "Cascadia Mono",
			Size:   11,
			Color:  "189855",
		},
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})

	monthStep = columnOffset
	amountEnd := rowOffset + len(categories)

	file.SetCellValue(ft, fmt.Sprintf("D%d", amountEnd), "Total")
	file.SetCellStyle(ft, fmt.Sprintf("D%d", amountEnd), fmt.Sprintf("D%d", amountEnd), categoryStyles)

	for i := 0; i < len(finances)+1; i++ {
		file.SetCellStyle(ft, fmt.Sprintf("%v%d", string(monthStep), amountEnd), fmt.Sprintf("%v%d", string(monthStep), amountEnd), bottomStyles)
		file.SetCellFormula(ft, fmt.Sprintf("%s%d", string(monthStep), amountEnd), fmt.Sprintf("SUM(%s9:%s%d)", string(monthStep), string(monthStep), amountEnd-1))
		monthStep++
	}

	// Chart
	var series []excelize.ChartSeries
	monthEnd := columnOffset + len(finances) - 1

	for j := rowOffset; j < len(categories)+rowOffset; j++ {
		series = append(series, excelize.ChartSeries{
			Name:       fmt.Sprintf("%s!$D$%d", ft, j),
			Categories: fmt.Sprintf("%s!$E$8:$%s$8", ft, string(monthEnd)),
			Values:     fmt.Sprintf("%s!$E$%d:$%s$%d", ft, j, string(monthEnd), j),
		})

		file.SetCellFormula(ft, fmt.Sprintf("%s%d", string(monthEnd+1), j), fmt.Sprintf("SUM(E%d:%s%d)", j, string(monthEnd), j))
		file.SetCellStyle(ft, fmt.Sprintf("%s%d", string(monthEnd+1), j), fmt.Sprintf("%s%d", string(monthEnd+1), j), bottomStyles)
	}

	if err := file.AddChart(ft, fmt.Sprintf("%c8", monthEnd+3), &excelize.Chart{
		Type:   excelize.Col3DClustered,
		Series: series,
		Title:  excelize.ChartTitle{Name: "Yearly report of " + strings.Title(ft) + "s"},
		XAxis:  excelize.ChartAxis{Font: excelize.Font{Color: "#000000"}},
		YAxis:  excelize.ChartAxis{Font: excelize.Font{Color: "#000000"}},
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
