package service

import (
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

type Employee struct {
	EmployeeID     string
	LastName       string // 姓
	FirstName      string // 名
	NickName       string // 表示名
	Role           string // 権限
	Password       string // パスワード
	DepartmentName string // 部署
	PositionName   string // 役職
	Phone          string // 電話番号
	Email          string // メールアドレス
}

func AddDataToExcel(employees []Employee, departments, positions []string) (string, error) {
	templateFileName := "employee_list.xlsx"

	var f *excelize.File

	// Check if file exists
	if _, err := os.Stat(templateFileName); os.IsNotExist(err) {
		// Create a new file if the template doesn't exist
		f = excelize.NewFile()
		f.NewSheet("従業員")
		f.NewSheet("部署") // Departments
		f.NewSheet("役職") // Positions
	} else {
		// Open the existing template file
		f, err = excelize.OpenFile(templateFileName)
		if err != nil {
			return "", fmt.Errorf("failed to open template file: %w", err)
		}
	}
	defer f.Close()

	// Write Employee Data to the "Employees" sheet
	employeeSheet := "従業員"
	f.SetSheetName("Sheet1", employeeSheet)
	headers := []string{"社員番号", "姓", "名", "表示名", "権限", "パスワード", "部署", "役職", "電話番号", "メールアドレス"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		if err := f.SetCellValue(employeeSheet, cell, header); err != nil {
			return "", fmt.Errorf("failed to write header in Employees sheet: %w", err)
		}
	}

	for i, emp := range employees {
		row := i + 2 // Start from the second row
		values := []interface{}{emp.EmployeeID, emp.LastName, emp.FirstName, emp.NickName, emp.Role, emp.Password, emp.DepartmentName, emp.PositionName, emp.Phone, emp.Email}
		for j, value := range values {
			cell := fmt.Sprintf("%c%d", 'A'+j, row)
			if err := f.SetCellValue(employeeSheet, cell, value); err != nil {
				return "", fmt.Errorf("failed to write employee data: %w", err)
			}
		}
	}

	// Write Department Data to the "部署" sheet
	departmentSheet := "部署"
	for i, dept := range departments {
		cell := fmt.Sprintf("A%d", i+2) // Start from the first row
		if err := f.SetCellValue(departmentSheet, cell, dept); err != nil {
			return "", fmt.Errorf("failed to write department data: %w", err)
		}
	}

	// Write Position Data to the "役職" sheet
	positionSheet := "役職"
	for i, pos := range positions {
		cell := fmt.Sprintf("A%d", i+2) // Start from the first row
		if err := f.SetCellValue(positionSheet, cell, pos); err != nil {
			return "", fmt.Errorf("failed to write position data: %w", err)
		}
	}

	// Save the file
	if err := f.SaveAs(templateFileName); err != nil {
		return "", fmt.Errorf("failed to save the Excel file: %w", err)
	}

	return "employee_list.xlsx", nil

}

func SaveInvalidUsersToExcel(employees []Employee, departments, positions []string) (string, error) {
	templateFileName := "employee_list.xlsx"

	var f *excelize.File

	// Check if file exists
	if _, err := os.Stat(templateFileName); os.IsNotExist(err) {
		// Create a new file if the template doesn't exist
		f = excelize.NewFile()
		f.NewSheet("従業員")
		f.NewSheet("部署") // Departments
		f.NewSheet("役職") // Positions
	} else {
		// Open the existing template file
		f, err = excelize.OpenFile(templateFileName)
		if err != nil {
			return "", fmt.Errorf("failed to open template file: %w", err)
		}
	}
	defer f.Close()

	// Write Employee Data to the "Employees" sheet
	employeeSheet := "従業員"
	f.SetSheetName("Sheet1", employeeSheet)
	headers := []string{"社員番号", "姓", "名", "表示名", "権限", "パスワード", "部署", "役職", "電話番号", "メールアドレス"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		if err := f.SetCellValue(employeeSheet, cell, header); err != nil {
			return "", fmt.Errorf("failed to write header in Employees sheet: %w", err)
		}
	}

	for i, emp := range employees {
		row := i + 2 // Start from the second row
		values := []interface{}{emp.EmployeeID, emp.LastName, emp.FirstName, emp.NickName, emp.Role, emp.Password, emp.DepartmentName, emp.PositionName, emp.Phone, emp.Email}
		for j, value := range values {
			cell := fmt.Sprintf("%c%d", 'A'+j, row)
			if err := f.SetCellValue(employeeSheet, cell, value); err != nil {
				return "", fmt.Errorf("failed to write employee data: %w", err)
			}
		}
	}

	// Write Department Data to the "部署" sheet
	departmentSheet := "部署"
	for i, dept := range departments {
		cell := fmt.Sprintf("A%d", i+2) // Start from the first row
		if err := f.SetCellValue(departmentSheet, cell, dept); err != nil {
			return "", fmt.Errorf("failed to write department data: %w", err)
		}
	}

	// Write Position Data to the "役職" sheet
	positionSheet := "役職"
	for i, pos := range positions {
		cell := fmt.Sprintf("A%d", i+2) // Start from the first row
		if err := f.SetCellValue(positionSheet, cell, pos); err != nil {
			return "", fmt.Errorf("failed to write position data: %w", err)
		}
	}

	// Save the file
	if err := f.SaveAs(templateFileName); err != nil {
		return "", fmt.Errorf("failed to save the Excel file: %w", err)
	}

	return "employee_list.xlsx", nil

}
