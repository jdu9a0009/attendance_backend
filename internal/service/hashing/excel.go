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
	// sheetName := "Sheet1"
	// Open the Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		// Close the file to release resources
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Read rows from the specified sheet
	// rows, err := f.GetRows(sheetName)
	// if err != nil {
	// 	return nil, err
	// }

	var users []UserExcellData

	// Iterate through the rows, starting from the second row to skip headers
	// for i, row := range rows {
	// 	if i == 0 {
	// 		// Skip the header row
	// 		continue
	// 	}

	// 	// var user UserExcellData
	// 	// for j, colCell := range row {
	// 	// 	switch j {
	// 	// 	case 0:
	// 	// 	case 1:

	// 	// 	}
	// 	// }
	// 	// users = append(users, user)
	// }

	return users, nil
}
