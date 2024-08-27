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

	// POST watermark
	req, err := http.NewRequest("POST", "http://localhost:8022/generate-excel", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	var resByte []byte
	resByte, err = io.ReadAll(resp.Body)
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
	EmployeeID   string
	Password     string
	Role         string
	FullName     string
	DepartmentID int
	PositionID   int
	Phone        string
	Email        string
}

func ExcelReader(filePath string) ([]UserExcellData, error) {
	// Open the Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer func() {
		// Close the file to release resources
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Get the sheet name
	sheetName := "Sheet1" // Replace with your actual sheet name

	// Get the total number of rows in the sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows from sheet: %w", err)
	}

	// Read data from each row
	var excelData []UserExcellData
	for i, row := range rows {
		if i == 0 { // Skip the header row
			continue
		}

		// Extract data from each cell in the row
		var data UserExcellData
		data.EmployeeID = row[0]
		data.Password = row[1]
		data.Role = row[2]
		data.FullName = row[3]
		// ... (extract other fields)

		excelData = append(excelData, data)
	}

	return excelData, nil
}
