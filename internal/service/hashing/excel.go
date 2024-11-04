package hashing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	Password       string
	Role           string
	FullName       string
	DepartmentName string
	PositionName   string
	Phone          string
	Email          string
}

func ExcelReader(filePath string,  fields map[int]string) ([]UserExcellData, []int, error) {
	sheetName:="Employee"
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, err
	}

	var users []UserExcellData
	var incompleteRows []int

	for i, row := range rows {
		if i == 0 {
			continue
		}

		var user UserExcellData
		isComplete := true

		for colIdx, fieldName := range fields {
			if colIdx < len(row) {
				value := row[colIdx]
				if value == "" {
					isComplete = false
				}

				switch fieldName {
				case "EmployeeID":
					user.EmployeeID = value
				case "Password":
					user.Password = value
				case "Role":
					user.Role = value
				case "FullName":
					user.FullName = value
				case "DepartmentName":
					user.DepartmentName = value
				case "PositionName":
					user.PositionName = value
				case "Phone":
					user.Phone = value
				case "Email":
					user.Email = value
				}
			}
		}

		if isComplete {
			users = append(users, user)
		} else {
			incompleteRows = append(incompleteRows, i+1)
		}
	}

	return users, incompleteRows, nil
}
