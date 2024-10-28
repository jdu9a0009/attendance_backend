package hashing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

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

func ExcelReader(filePath string) ([]UserExcellData, []int, error) {
	sheetName := "Sheet1"
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
		isComplete := true // Track if row is complete

		for j, colCell := range row {
			switch j {
			case 1:
				user.Password = colCell
				if colCell == "" {
					isComplete = false
				}
			case 2:
				user.Role = colCell
				if colCell == "" {
					isComplete = false
				}
			case 3:
				user.FullName = colCell
				if colCell == "" {
					isComplete = false
				}
			case 4:
				if colCell != "" {
					departmentID, err := strconv.Atoi(colCell)
					if err != nil {
						return nil, nil, fmt.Errorf("invalid department ID in row %d: %v", i+1, err)
					}
					user.DepartmentID = departmentID
				} else {
					isComplete = false
				}
			case 5:
				if colCell != "" {
					positionID, err := strconv.Atoi(colCell)
					if err != nil {
						return nil, nil, fmt.Errorf("invalid position ID in row %d: %v", i+1, err)
					}
					user.PositionID = positionID
				} else {
					isComplete = false
				}
			case 6:
				user.Phone = colCell
				if colCell == "" {
					isComplete = false
				}
			case 7:
				user.Email = colCell
				if colCell == "" {
					isComplete = false
				}
			}
		}

		if isComplete {
			users = append(users, user)
		} else {
			incompleteRows = append(incompleteRows, i+1) // Store the Excel row number
		}
	}

	return users, incompleteRows, nil
}
