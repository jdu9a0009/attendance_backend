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

func ExcelReaderByCreate(filePath string, fields map[int]string, departmentMap, positionMap map[string]int, employeeIDList []string) ([]UserExcellData, []int, error) {
	sheetName := "Employees"
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

	// Regex patterns for email and phone validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex := regexp.MustCompile(`^\+?\d*$`)

	existingEmployeeIDs := make(map[string]struct{})
	for _, id := range employeeIDList {
		existingEmployeeIDs[id] = struct{}{}
	}

	var users []UserExcellData
	var incompleteRows []int
	localEmployeeIDSet := make(map[string]int) // To track duplicate IDs within the file

	for i, row := range rows {
		if i == 0 {
			// Skip the header row
			continue
		}

		if len(row) < 10 { // Adjust the number based on expected columns
			incompleteRows = append(incompleteRows, i)
			continue
		}

		// Trim spaces
		employeeID := strings.TrimSpace(row[0])
		lastName := strings.TrimSpace(row[1])
		firstName := strings.TrimSpace(row[2])
		role := strings.TrimSpace(row[4])
		password := strings.TrimSpace(row[5])
		department := strings.TrimSpace(row[6])
		position := strings.TrimSpace(row[7])
		email := strings.TrimSpace(row[9])
		phone := strings.TrimSpace(row[8])

		// Check mandatory fields
		if employeeID == "" || lastName == "" || firstName == "" || role == "" || password == "" || department == "" || position == "" {
			incompleteRows = append(incompleteRows, i)

		}

		// Check department and position mapping
		departmentID, okDept := departmentMap[department]
		positionID, okPos := positionMap[position]
		if !okDept || !okPos {
			incompleteRows = append(incompleteRows, i)
			continue
		}

		// Check if employee ID contains only half-width characters
		if !isHalfWidth(employeeID) {
			incompleteRows = append(incompleteRows, i) // Invalid employee ID format
			continue
		}

		// Check if employee ID is already in the local or global set
		if _, exists := existingEmployeeIDs[employeeID]; exists || localEmployeeIDSet[employeeID] != 0 {
			incompleteRows = append(incompleteRows, i) // Duplicate employee ID
			continue
		}

		// Validate email if provided
		if email != "" && !emailRegex.MatchString(email) {
			incompleteRows = append(incompleteRows, i) // Invalid email format
			continue
		}

		// Validate phone if provided
		if phone != "" && !phoneRegex.MatchString(phone) {
			incompleteRows = append(incompleteRows, i) // Invalid phone format
			continue
		}

		// Add the employee ID to the local set
		localEmployeeIDSet[employeeID] = i

		// Append the user data
		users = append(users, UserExcellData{
			EmployeeID:   employeeID,
			LastName:     lastName,
			FirstName:    firstName,
			NickName:     row[3],
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

func ExcelReaderByEdit(filePath string, fields map[int]string, departmentMap, positionMap map[string]int) ([]UserExcellData, []int, error) {
	sheetName := "Employees"
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

		// Check if the row has fewer columns than required
		if len(row) < 10 {
			incompleteRows = append(incompleteRows, i)
			continue
		}

		// Map department and position
		departmentID, okDept := departmentMap[row[6]]
		positionID, okPos := positionMap[row[7]]
		if !okDept || !okPos {
			incompleteRows = append(incompleteRows, i)
			continue
		}

		// Add user data to the users slice
		users = append(users, UserExcellData{
			EmployeeID:   row[0],
			LastName:     row[1],
			FirstName:    row[2],
			NickName:     row[3],
			Role:         row[4],
			Password:     row[5],
			DepartmentID: departmentID,
			PositionID:   positionID,
			Phone:        row[8],
			Email:        row[9],
		})
	}

	return users, incompleteRows, nil
}

func ExcelReaderByDelete(filePath string, rowLen int, fields map[int]string) ([]string, []int, error) {
	sheetName := "Employees"
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
