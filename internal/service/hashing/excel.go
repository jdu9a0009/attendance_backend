package hashing

import (
	"attendance/backend/foundation/web"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/unicode/norm"
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
	LastName       string
	FirstName      string
	NickName       string
	Role           string
	Password       string
	DepartmentName string
	DepartmentID   int
	PositionName   string
	PositionID     int
	Phone          string
	Email          string
}

func ExcelReaderByCreate(filePath string, fields map[int]string, departmentMap, positionMap map[string]int, employeeIDMap, existingEmailMap map[string]struct{}) ([]UserExcellData, []int, error) {
	sheetName := "従業員"
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, err
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex := regexp.MustCompile(`^\+?\d+$`)

	var users []UserExcellData
	var incompleteRows []int
	localEmployeeIDs := make(map[string]int) // Track IDs in current file
	localEmails := make(map[string]int)      // Track emails in current file

	for i, row := range rows {
		if i == 0 {
			continue // Skip header
		}

		if len(row) < 10 {
			incompleteRows = append(incompleteRows, i+1)
			continue
		}

		// Trim and validate fields
		employeeID := strings.TrimSpace(row[0])
		lastName := strings.TrimSpace(row[1])
		firstName := strings.TrimSpace(row[2])
		role := strings.TrimSpace(row[4])
		password := strings.TrimSpace(row[5])
		department := strings.TrimSpace(row[6])
		position := strings.TrimSpace(row[7])
		phone := strings.TrimSpace(row[8])
		email := strings.TrimSpace(row[9])

		// Mandatory fields check
		if employeeID == "" || lastName == "" || firstName == "" ||
			role == "" || password == "" || department == "" || position == "" {
			incompleteRows = append(incompleteRows, i+1)
			continue
		}

		// Half-width characters check
		if !isHalfWidth(employeeID) || !isHalfWidth(password) ||
			(email != "" && !isHalfWidth(email)) {
			incompleteRows = append(incompleteRows, i+1)
			continue
		}

		// Employee ID uniqueness checks
		if _, exists := employeeIDMap[employeeID]; exists {
			incompleteRows = append(incompleteRows, i+1) // Existing in DB
			continue
		}
		if prevRow, exists := localEmployeeIDs[employeeID]; exists {
			incompleteRows = append(incompleteRows, prevRow, i+1)
			continue
		}

		// Email uniqueness checks
		if email != "" {
			if _, exists := existingEmailMap[email]; exists {
				incompleteRows = append(incompleteRows, i+1) // Existing in DB
				continue
			}
			if prevRow, exists := localEmails[email]; exists {
				incompleteRows = append(incompleteRows, prevRow, i+1)
				continue
			}
		}

		// Department/Position validation
		departmentID, deptOK := departmentMap[department]
		positionID, posOK := positionMap[position]
		if !deptOK || !posOK {
			incompleteRows = append(incompleteRows, i+1)
			continue
		}

		// Email format validation
		if email != "" && !emailRegex.MatchString(email) {
			incompleteRows = append(incompleteRows, i+1)
			continue
		}

		// Phone format validation
		if phone != "" && !phoneRegex.MatchString(phone) {
			incompleteRows = append(incompleteRows, i+1)
			continue
		}

		// Track unique values
		localEmployeeIDs[employeeID] = i + 1
		if email != "" {
			localEmails[email] = i + 1
		}

		users = append(users, UserExcellData{
			EmployeeID:   employeeID,
			LastName:     lastName,
			FirstName:    firstName,
			NickName:     strings.TrimSpace(row[3]),
			Role:         role,
			Password:     password,
			DepartmentID: departmentID,
			PositionID:   positionID,
			Phone:        phone,
			Email:        email,
		})
	}

	return users, incompleteRows, nil
}

func ExcelReaderByEdit(filePath string, fields map[int]string, departmentMap, positionMap map[string]int, existingIDs, existingEmails map[string]struct{}) ([]UserExcellData, []int, error) {
	sheetName := "従業員"
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Printf("File close error: %v", closeErr)
		}
	}()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, err
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex := regexp.MustCompile(`^\+?\d+$`)

	var users []UserExcellData
	var incompleteRows []int
	localIDs := make(map[string]int)
	localEmails := make(map[string]int)

	for i, row := range rows {
		rowNumber := i + 1 // Excel 1-based row numbers
		if i == 0 {
			continue // Skip header
		}

		// Check minimum required columns
		if len(row) < 10 {
			incompleteRows = append(incompleteRows, rowNumber)
			continue
		}

		// Trim all fields
		employeeID := strings.TrimSpace(row[0])
		lastName := strings.TrimSpace(row[1])
		firstName := strings.TrimSpace(row[2])
		role := strings.TrimSpace(row[4])
		password := strings.TrimSpace(row[5])
		department := strings.TrimSpace(row[6])
		position := strings.TrimSpace(row[7])
		phone := strings.TrimSpace(row[8])
		email := strings.TrimSpace(row[9])

		// Validate mandatory fields
		if employeeID == "" || lastName == "" || firstName == "" ||
			role == "" || department == "" || position == "" {
			incompleteRows = append(incompleteRows, rowNumber)
			continue
		}

		// Validate half-width characters
		if !isHalfWidth(employeeID) || (password != "" && !isHalfWidth(password)) ||
			(email != "" && !isHalfWidth(email)) {
			incompleteRows = append(incompleteRows, rowNumber)
			continue
		}

		// Check department and position existence
		departmentID, deptOK := departmentMap[department]
		positionID, posOK := positionMap[position]
		if !deptOK || !posOK {
			incompleteRows = append(incompleteRows, rowNumber)
			continue
		}

		// Email validation
		if email != "" {
			if !emailRegex.MatchString(email) {
				incompleteRows = append(incompleteRows, rowNumber)
				continue
			}
		}

		// Phone validation
		if phone != "" && !phoneRegex.MatchString(phone) {
			incompleteRows = append(incompleteRows, rowNumber)
			continue
		}

		// Check local duplicates
		if prevRow, exists := localIDs[employeeID]; exists {
			incompleteRows = append(incompleteRows, prevRow, rowNumber)
			continue
		}
		if email != "" {
			if prevRow, exists := localEmails[email]; exists {
				incompleteRows = append(incompleteRows, prevRow, rowNumber)
				continue
			}
		}

		// Check against existing IDs and Emails (global duplicates)
		if _, exists := existingIDs[employeeID]; exists {
			incompleteRows = append(incompleteRows, rowNumber)
			continue
		}
		if email != "" {
			if _, exists := existingEmails[email]; exists {
				incompleteRows = append(incompleteRows, rowNumber)
				continue
			}
		}

		// Track processed values
		localIDs[employeeID] = rowNumber
		if email != "" {
			localEmails[email] = rowNumber
		}

		users = append(users, UserExcellData{
			EmployeeID:   employeeID,
			LastName:     lastName,
			FirstName:    firstName,
			NickName:     strings.TrimSpace(row[3]),
			Role:         role,
			Password:     password,
			DepartmentID: departmentID,
			PositionID:   positionID,
			Phone:        phone,
			Email:        email,
		})
	}

	return users, incompleteRows, nil
}

func ExcelReaderByDelete(filePath string, rowLen int, fields map[int]string) ([]string, []int, error) {
	sheetName := "従業員"
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

	var employeeIDs []string
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

		// Collect only EmployeeID (column 0 in your data)
		if len(row) > 0 && row[0] != "" {
			employeeIDs = append(employeeIDs, row[0])
		} else {
			incompleteRows = append(incompleteRows, i)
		}
	}
	return employeeIDs, incompleteRows, nil
}

// isHalfWidth checks if a string contains only half-width characters.
func isHalfWidth(s string) bool {
	// Normalize the string to NFC form.
	normalized := norm.NFC.String(s)
	for _, r := range normalized {
		// Full-width character detection
		if r >= '\uFF01' && r <= '\uFF60' || r >= '\uFFE0' && r <= '\uFFEF' {
			return false
		}
	}
	return true
}

func ValidateHalfWidthInput() web.Middleware {
	return func(handler web.Handler) web.Handler {
		return func(c *web.Context) error {
			// Iterate over form values and validate each one.
			for _, values := range c.Request.Form {
				for _, value := range values {
					if !isHalfWidth(value) {
						return c.RespondError(web.NewRequestError(
							errors.New("入力は半角文字のみ使用可能"), http.StatusBadRequest))
					}
				}
			}

			// Proceed to the next handler if validation passes.
			return handler(c)
		}
	}
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
