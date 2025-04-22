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
		rowNumber := i + 1 // Excel qator raqami (1-based)
		if i == 0 {
			continue // Header qatorini o'tkazib yuborish
		}

		// Yetarlicha ustun mavjudligini tekshirish
		if len(row) < 10 {
			log.Printf("Row %d: Missing columns. Expected at least 10, got %d\n", rowNumber, len(row))
			incompleteRows = append(incompleteRows, rowNumber)
			continue
		}

		// Qator ma'lumotlarini olish
		employeeID := strings.TrimSpace(row[0])
		lastName := strings.TrimSpace(row[1])
		firstName := strings.TrimSpace(row[2])
		role := strings.TrimSpace(row[4])
		password := strings.TrimSpace(row[5])
		department := strings.TrimSpace(row[6])
		position := strings.TrimSpace(row[7])
		phone := strings.TrimSpace(row[8])
		email := strings.TrimSpace(row[9])

		// Majburiy ustunlarni tekshirish
		missingColumns := []string{}
		if employeeID == "" {
			missingColumns = append(missingColumns, "Employee ID")
		}
		if lastName == "" {
			missingColumns = append(missingColumns, "Last Name")
		}
		if firstName == "" {
			missingColumns = append(missingColumns, "First Name")
		}
		if role == "" {
			missingColumns = append(missingColumns, "Role")
		}
		if department == "" {
			missingColumns = append(missingColumns, "Department")
		}
		if position == "" {
			missingColumns = append(missingColumns, "Position")
		}

		if len(missingColumns) > 0 {
			log.Printf("Row %d: Missing required columns: %v\n", rowNumber, missingColumns)
			incompleteRows = append(incompleteRows, rowNumber)
			continue
		}
		// Half-width characters check
		if !isHalfWidth(employeeID) || !isHalfWidth(password) ||
			(email != "" && !isHalfWidth(email)) {
			incompleteRows = append(incompleteRows, i+1)
			continue
		}
		// Department va Position tekshirish
		departmentID, deptOK := departmentMap[department]
		positionID, posOK := positionMap[position]
		if !deptOK || !posOK {
			log.Printf("Row %d: Invalid department or position - Department: '%s', Position: '%s'\n", rowNumber, department, position)
			incompleteRows = append(incompleteRows, rowNumber)
			continue
		}

		// Email tekshirish
		if email != "" && !emailRegex.MatchString(email) {
			log.Printf("Row %d: Invalid email format: %s\n", rowNumber, email)
			incompleteRows = append(incompleteRows, rowNumber)
			continue
		}

		// Telefon raqamini tekshirish
		if phone != "" && !phoneRegex.MatchString(phone) {
			log.Printf("Row %d: Invalid phone format: %s\n", rowNumber, phone)
			incompleteRows = append(incompleteRows, rowNumber)
			continue
		}
		// Check local duplicates
		if prevRow, exists := localIDs[employeeID]; exists {
			log.Printf("Row %d: Duplicate Employee ID '%s' found (previously seen in Row %d)\n", rowNumber, employeeID, prevRow)
			incompleteRows = append(incompleteRows, rowNumber) // Faqat joriy qatorni qo'shish
			continue
		}
		if email != "" {
			if prevRow, exists := localEmails[email]; exists {
				log.Printf("Row %d: Duplicate Email '%s' found (previously seen in Row %d)\n", rowNumber, email, prevRow)
				incompleteRows = append(incompleteRows, rowNumber) // Faqat joriy qatorni qo'shish
				continue
			}
		}

		// Check against existing IDs and Emails (global duplicates)
		// if _, exists := existingIDs[employeeID]; exists {
		// 	log.Printf("Row %d: Employee ID '%s' already exists in the system\n", rowNumber, employeeID)
		// 	incompleteRows = append(incompleteRows, rowNumber)
		// 	continue
		// }
		// if email != "" {
		// 	if _, exists := existingEmails[email]; exists {
		// 		log.Printf("Row %d: Email '%s' already exists in the system\n", rowNumber, email)
		// 		incompleteRows = append(incompleteRows, rowNumber)
		// 		continue
		// 	}
		// }
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

// func SaveInvalidUsersToExcel(employees []UserExcellData) (string, error) {
// 	templateFileName := "invalid_employees.xlsx"

// 	var f *excelize.File

// 	// Check if file exists
// 	if _, err := os.Stat(templateFileName); os.IsNotExist(err) {
// 		// Create a new file if the template doesn't exist
// 		f = excelize.NewFile()
// 		f.NewSheet("従業員")
// 	} else {
// 		// Open the existing template file
// 		f, err = excelize.OpenFile(templateFileName)
// 		if err != nil {
// 			return "", fmt.Errorf("failed to open template file: %w", err)
// 		}
// 	}
// 	defer f.Close()

// 	// Write Employee Data to the "Employees" sheet
// 	employeeSheet := "従業員"
// 	f.SetSheetName("Sheet1", employeeSheet)
// 	headers := []string{"社員番号", "姓", "名", "表示名", "権限", "パスワード", "部署", "役職", "電話番号", "メールアドレス", "エラー"}
// 	for i, header := range headers {
// 		cell := fmt.Sprintf("%c1", 'A'+i)
// 		if err := f.SetCellValue(employeeSheet, cell, header); err != nil {
// 			return "", fmt.Errorf("failed to write header in Employees sheet: %w", err)
// 		}
// 	}

// 	for i, emp := range employees {
// 		row := i + 2 // Start from the second row
// 		values := []interface{}{emp.EmployeeID, emp.LastName, emp.FirstName, emp.NickName, emp.Role, emp.Password, emp.DepartmentName, emp.PositionName, emp.Phone, emp.Email, emp.Error}
// 		for j, value := range values {
// 			cell := fmt.Sprintf("%c%d", 'A'+j, row)
// 			if err := f.SetCellValue(employeeSheet, cell, value); err != nil {
// 				return "", fmt.Errorf("failed to write employee data: %w", err)
// 			}
// 		}
// 	}

// 	// Save the file
// 	if err := f.SaveAs(templateFileName); err != nil {
// 		return "", fmt.Errorf("failed to save the Excel file: %w", err)
// 	}

// 	return "invalid_employees.xlsx", nil

// }
