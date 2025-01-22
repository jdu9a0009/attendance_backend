package service

import (
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

type Employee struct {
	EmployeeID     string
	FirstName      string
	LastName       string
	NickName       string
	Role           string
	DepartmentName string
	PositionName   string
	Phone          string
	Email          string
}

func AddDataToExcel(employees []Employee, fileName string) error {
	var f *excelize.File
	// var err error
	sheet := "Sheet1"
	// Check if file exists
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		// File does not exist, create a new one
		f = excelize.NewFile()
		f.SetSheetName("Sheet1", sheet)

		// Set headers in the first row
		headers := []string{"Employee ID", "FirstName", "LastName", "NickName", "Department Name", "Position Name", "Phone Number", "Email"}
		for i, header := range headers {
			cell := fmt.Sprintf("%c1", 'A'+i)
			f.SetCellValue(sheet, cell, header)
		}
	} else {
		// File exists, open it
		f, err = excelize.OpenFile(fileName)
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}
	}

	// Find the next empty row
	rowNum := 2
	for {
		cell, _ := f.GetCellValue(sheet, fmt.Sprintf("A%d", rowNum))
		if cell == "" {
			break
		}
		rowNum++
	}

	// Populate rows with data starting from the next empty row
	for _, entry := range employees {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowNum), entry.EmployeeID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowNum), entry.FirstName)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowNum), entry.LastName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowNum), entry.NickName)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", rowNum), entry.DepartmentName)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", rowNum), entry.PositionName)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", rowNum), entry.Phone)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", rowNum), entry.Email)
		rowNum++
	}

	// Save the file
	if err := f.SaveAs(fileName); err != nil {
		return fmt.Errorf("error saving file: %w", err)
	}
	return nil
}
