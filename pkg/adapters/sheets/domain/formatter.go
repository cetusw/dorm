package domain

import (
	"google.golang.org/api/sheets/v4"
)

type SheetFormatter struct {
	sheetID int64
}

func NewSheetFormatter(sheetID int64) *SheetFormatter {
	return &SheetFormatter{sheetID: sheetID}
}

func (f *SheetFormatter) NewRange(
	rowStart int64,
	rowEnd int64,
	columnStart int64,
	columnEnd int64,
) *sheets.GridRange {
	return &sheets.GridRange{
		SheetId:          f.sheetID,
		StartRowIndex:    rowStart,
		EndRowIndex:      rowEnd,
		StartColumnIndex: columnStart,
		EndColumnIndex:   columnEnd,
	}
}

func (f *SheetFormatter) UpdateBorders(
	gridRange *sheets.GridRange,
	top bool,
	bottom bool,
	left bool,
	right bool,
	innerH bool,
	innerV bool,
) *sheets.Request {
	border := &sheets.Border{Style: "SOLID", Color: &sheets.Color{Red: 0, Green: 0, Blue: 0}}
	req := &sheets.UpdateBordersRequest{Range: gridRange}
	if top {
		req.Top = border
	}
	if bottom {
		req.Bottom = border
	}
	if left {
		req.Left = border
	}
	if right {
		req.Right = border
	}
	if innerH {
		req.InnerHorizontal = border
	}
	if innerV {
		req.InnerVertical = border
	}
	return &sheets.Request{UpdateBorders: req}
}

func (f *SheetFormatter) RepeatCell(
	gridRange *sheets.GridRange,
	format *sheets.CellFormat,
	fields string,
) *sheets.Request {
	return &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range:  gridRange,
			Cell:   &sheets.CellData{UserEnteredFormat: format},
			Fields: fields,
		},
	}
}

func (f *SheetFormatter) MergeCells(
	rowStart int64,
	rowEnd int64,
	columnStart int64,
	columnEnd int64,
	mType string,
) *sheets.Request {
	return &sheets.Request{
		MergeCells: &sheets.MergeCellsRequest{
			Range:     f.NewRange(rowStart, rowEnd, columnStart, columnEnd),
			MergeType: mType,
		},
	}
}

func (f *SheetFormatter) SetDimensionSize(
	dimension string,
	index int64,
	size int64,
) *sheets.Request {
	return &sheets.Request{
		UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
			Range:      &sheets.DimensionRange{SheetId: f.sheetID, Dimension: dimension, StartIndex: index, EndIndex: index + 1},
			Properties: &sheets.DimensionProperties{PixelSize: size},
			Fields:     "pixelSize",
		},
	}
}
