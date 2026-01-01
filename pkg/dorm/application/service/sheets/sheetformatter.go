package sheets

import (
	"fmt"
	"strings"

	"google.golang.org/api/sheets/v4"
)

type SheetFormatter struct {
	sheetID int64
}

func NewSheetFormatter(sheetID int64) *SheetFormatter {
	return &SheetFormatter{sheetID: sheetID}
}

func (f *SheetFormatter) SetColumnWidth(colIndex int64, width int64) *sheets.Request {
	return &sheets.Request{
		UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
			Range: &sheets.DimensionRange{
				SheetId:    f.sheetID,
				Dimension:  "COLUMNS",
				StartIndex: colIndex,
				EndIndex:   colIndex + 1,
			},
			Properties: &sheets.DimensionProperties{
				PixelSize: width,
			},
			Fields: "pixelSize",
		},
	}
}

func (f *SheetFormatter) SetRowHeight(rowIndex int64, height int64) *sheets.Request {
	return &sheets.Request{
		UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
			Range: &sheets.DimensionRange{
				SheetId:    f.sheetID,
				Dimension:  "ROWS",
				StartIndex: rowIndex,
				EndIndex:   rowIndex + 1,
			},
			Properties: &sheets.DimensionProperties{
				PixelSize: height,
			},
			Fields: "pixelSize",
		},
	}
}

func (f *SheetFormatter) MergeCells(startRow, endRow, startCol, endCol int64, mergeType string) *sheets.Request {
	return &sheets.Request{
		MergeCells: &sheets.MergeCellsRequest{
			Range: &sheets.GridRange{
				SheetId:          f.sheetID,
				StartRowIndex:    startRow,
				EndRowIndex:      endRow,
				StartColumnIndex: startCol,
				EndColumnIndex:   endCol,
			},
			MergeType: mergeType, // "MERGE_ALL", "MERGE_COLUMNS", "MERGE_ROWS"
		},
	}
}

func (f *SheetFormatter) SetBackgroundColor(startRow, endRow, startCol, endCol int64, color *sheets.Color) *sheets.Request {
	return &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          f.sheetID,
				StartRowIndex:    startRow,
				EndRowIndex:      endRow,
				StartColumnIndex: startCol,
				EndColumnIndex:   endCol,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					BackgroundColor: color,
				},
			},
			Fields: "userEnteredFormat(backgroundColor)",
		},
	}
}

func (f *SheetFormatter) SetTextFormat(
	startRow, endRow, startCol, endCol int64,
	bold bool,
	fontSize int64,
	fontFamily string,
	textColor *sheets.Color,
) *sheets.Request {
	format := &sheets.TextFormat{}
	format.Bold = bold
	format.FontSize = fontSize
	format.FontFamily = fontFamily
	if textColor != nil {
		format.ForegroundColor = textColor
	}
	return &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          f.sheetID,
				StartRowIndex:    startRow,
				EndRowIndex:      endRow,
				StartColumnIndex: startCol,
				EndColumnIndex:   endCol,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					TextFormat: format,
				},
			},
			Fields: "userEnteredFormat(textFormat)",
		},
	}
}

func (f *SheetFormatter) SetTextAlignment(
	startRow, endRow, startCol, endCol int64,
	hAlign, vAlign string, // "LEFT", "CENTER", "RIGHT" / "TOP", "MIDDLE", "BOTTOM"
) *sheets.Request {
	return &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          f.sheetID,
				StartRowIndex:    startRow,
				EndRowIndex:      endRow,
				StartColumnIndex: startCol,
				EndColumnIndex:   endCol,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					HorizontalAlignment: hAlign,
					VerticalAlignment:   vAlign,
				},
			},
			Fields: "userEnteredFormat(horizontalAlignment,verticalAlignment)",
		},
	}
}

func (f *SheetFormatter) SetWrapStrategy(startRow, endRow, startCol, endCol int64, wrap string) *sheets.Request {
	return &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          f.sheetID,
				StartRowIndex:    startRow,
				EndRowIndex:      endRow,
				StartColumnIndex: startCol,
				EndColumnIndex:   endCol,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					WrapStrategy: wrap, // "WRAP", "OVERFLOW_CELL", "CLIP"
				},
			},
			Fields: "userEnteredFormat(wrapStrategy)",
		},
	}
}

func (f *SheetFormatter) SetTextRotation(startRow, endRow, startCol, endCol int64, angle int64) *sheets.Request {
	return &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          f.sheetID,
				StartRowIndex:    startRow,
				EndRowIndex:      endRow,
				StartColumnIndex: startCol,
				EndColumnIndex:   endCol,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					TextRotation: &sheets.TextRotation{Angle: angle},
				},
			},
			Fields: "userEnteredFormat(textRotation)",
		},
	}
}

func (f *SheetFormatter) SetDataValidation(startRow, endRow, startCol, endCol int64, values []string) *sheets.Request {
	var conditionValues []*sheets.ConditionValue
	for _, v := range values {
		conditionValues = append(conditionValues, &sheets.ConditionValue{UserEnteredValue: v})
	}

	return &sheets.Request{
		SetDataValidation: &sheets.SetDataValidationRequest{
			Range: &sheets.GridRange{
				SheetId:          f.sheetID,
				StartRowIndex:    startRow,
				EndRowIndex:      endRow,
				StartColumnIndex: startCol,
				EndColumnIndex:   endCol,
			},
			Rule: &sheets.DataValidationRule{
				Condition: &sheets.BooleanCondition{
					Type:   "ONE_OF_LIST",
					Values: conditionValues,
				},
				ShowCustomUi: true,
				Strict:       true,
			},
		},
	}
}

func (f *SheetFormatter) AddConditionalFormatRule(startRow, endRow, startCol, endCol int64, rule *sheets.ConditionalFormatRule) *sheets.Request {
	rule.Ranges = []*sheets.GridRange{
		{
			SheetId:          f.sheetID,
			StartRowIndex:    startRow,
			EndRowIndex:      endRow,
			StartColumnIndex: startCol,
			EndColumnIndex:   endCol,
		},
	}

	return &sheets.Request{
		AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
			Rule:  rule,
			Index: 0,
		},
	}
}

func (f *SheetFormatter) SetOuterBorders(startRow, endRow, startCol, endCol int64) *sheets.Request {
	border := &sheets.Border{
		Style: "SOLID",
		ColorStyle: &sheets.ColorStyle{
			RgbColor: &sheets.Color{Red: 0, Green: 0, Blue: 0},
		},
	}

	return &sheets.Request{
		UpdateBorders: &sheets.UpdateBordersRequest{
			Range: &sheets.GridRange{
				SheetId:          f.sheetID,
				StartRowIndex:    startRow,
				EndRowIndex:      endRow,
				StartColumnIndex: startCol,
				EndColumnIndex:   endCol,
			},
			Top:    border,
			Bottom: border,
			Left:   border,
			Right:  border,
		},
	}
}

func (f *SheetFormatter) ClearFormatting(startRow, endRow, startCol, endCol int64) *sheets.Request {
	return &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          f.sheetID,
				StartRowIndex:    startRow,
				EndRowIndex:      endRow,
				StartColumnIndex: startCol,
				EndColumnIndex:   endCol,
			},
			Cell:   &sheets.CellData{},
			Fields: "userEnteredFormat",
		},
	}
}

func (f *SheetFormatter) ClearContent(startRow, endRow, startCol, endCol int64) *sheets.Request {
	return &sheets.Request{
		UpdateCells: &sheets.UpdateCellsRequest{
			Range: &sheets.GridRange{
				SheetId:          f.sheetID,
				StartRowIndex:    startRow,
				EndRowIndex:      endRow,
				StartColumnIndex: startCol,
				EndColumnIndex:   endCol,
			},
			Fields: "userEnteredValue",
		},
	}
}

func (f *SheetFormatter) SetValue(row, col int64, value interface{}) *sheets.Request {
	var extendedValue *sheets.ExtendedValue

	switch v := value.(type) {
	case string:
		extendedValue = &sheets.ExtendedValue{StringValue: &v}
	case int:
		f := float64(v)
		extendedValue = &sheets.ExtendedValue{NumberValue: &f}
	case int64:
		f := float64(v)
		extendedValue = &sheets.ExtendedValue{NumberValue: &f}
	case float64:
		extendedValue = &sheets.ExtendedValue{NumberValue: &v}
	case bool:
		extendedValue = &sheets.ExtendedValue{BoolValue: &v}
	default:
		s := fmt.Sprintf("%v", value)
		extendedValue = &sheets.ExtendedValue{StringValue: &s}
	}

	cellData := []*sheets.CellData{
		{UserEnteredValue: extendedValue},
	}

	return &sheets.Request{
		UpdateCells: &sheets.UpdateCellsRequest{
			Rows: []*sheets.RowData{
				{Values: cellData},
			},
			Start: &sheets.GridCoordinate{
				SheetId:     f.sheetID,
				RowIndex:    row,
				ColumnIndex: col,
			},
			Fields: "userEnteredValue",
		},
	}
}

func (f *SheetFormatter) IndexToLetter(index int64) string {
	if index < 0 {
		return ""
	}
	var result string
	for index >= 0 {
		result = string(rune('A'+index%26)) + result
		index = index/26 - 1
	}
	return result
}

func (f *SheetFormatter) LetterToIndex(letter string) (int64, error) {
	letter = strings.ToUpper(letter)
	var index int64 = 0
	for _, char := range letter {
		if char < 'A' || char > 'Z' {
			return 0, fmt.Errorf("invalid column letter: %s", letter)
		}
		index = index*26 + int64(char-'A'+1)
	}
	return index - 1, nil
}

func (f *SheetFormatter) GetID() int64 {
	return f.sheetID
}
