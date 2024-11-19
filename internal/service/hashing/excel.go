package hashing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
)

type ExcelData struct {
	Data struct {
		Keys   any `json:"keys"`
		Values any `json:"values"`
	} `json:"data"`
	BasePath  string `json:"base_path"`
	FileName  string `json:"file_name"`
	ExcelPath string `json:"excel_path"`
}

type response struct {
	Message string `json:"message"`
	Status  bool   `json:"status"`
	Data    *struct {
		Excel string `json:"excel"`
	} `json:"data"`
}

func ExcelDog(data ExcelData) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "http://localhost:8022/generate-excel", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	resByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	res := response{}
	err = json.Unmarshal(resByte, &res)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 || res.Data == nil {
		return "", errors.New(fmt.Sprintf("status code: %d and message: %s", resp.StatusCode, res.Message))
	}

	return res.Data.Excel, nil
}

type UserExcellData struct {
	EmployeeID     string
	FirstName      string
	LastName       string
	NickName       string
	Role           string
	Password       string
	DepartmentName string
	PositionName   string
	Phone          string
	Email          string
}

func ExcelReader(filePath string, rowLen int, fields map[int]string) ([]UserExcellData, []int, error) {
	sheetName := "Sheet1"
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, err
	}

	var users []UserExcellData
	var incompleteRows []int
	for i, row := range rows {
		if i == 0 {
			// Skip the header row
			continue
		}

		if len(row) < rowLen { // Check if the row has enough columns
			incompleteRows = append(incompleteRows, i)
			continue
		}

		var user UserExcellData

		// Only assign values if the column exists
		if len(row) > 0 { user.EmployeeID = row[0] }
		if len(row) > 1 { user.FirstName = row[1] }
		if len(row) > 2 { user.LastName = row[2] }
		if len(row) > 3 { user.NickName = row[3] }
		if len(row) > 4 { user.Role = row[4] }
		if len(row) > 5 { user.Password = row[5] }
		if len(row) > 6 { user.DepartmentName = row[6] }
		if len(row) > 7 { user.PositionName = row[7] }
		if len(row) > 8 { user.Phone = row[8] }
		if len(row) > 9 { user.Email = row[9] }

		users = append(users, user)
	}
	return users, incompleteRows, nil
}


func EditExcell(departments, positions []string) (string, error) {
	// Open the Excel file
	f, err := excelize.OpenFile("template.xlsx")
	if err != nil {
		return "", fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	department := "部署"
	position := "役職"

	// Check if the sheet exists
	if sheetIndex, err := f.GetSheetIndex(department); sheetIndex == -1 {
		if err != nil {
			return "", fmt.Errorf("failed to Department GetSheet  Excel file: %w", err)
		}
	}
	if sheetIndex, err := f.GetSheetIndex(position); sheetIndex == -1 {
		if err != nil {
			return "", fmt.Errorf("failed to Position GetSheet Excel file: %w", err)
		}
	}

	for i, dept := range departments {
		cell := fmt.Sprintf("A%d", i+2)
		if err := f.SetCellValue(department, cell, dept); err != nil {
			return "", fmt.Errorf("failed to write department data: %w", err)
		}
	}

	for i, pos := range positions {
		cell := fmt.Sprintf("A%d", i+2)
		if err := f.SetCellValue(position, cell, pos); err != nil {
			return "", fmt.Errorf("failed to write position data: %w", err)
		}
	}

	if err := f.Save(); err != nil {
		return "", fmt.Errorf("error saving file: %w", err)
	}
	return "template.xlsx", nil
}
