package service

import (
	"fmt"
	"time"

	"dorm/internal/dorm/infrastructure"
)

type CleaningService struct {
	sheets        *infrastructure.Sheets
	Telegram      *infrastructure.Telegram
	spreadsheetID string
}

func NewCleaningService(
	sheetsClient *infrastructure.Sheets,
	telegramClient *infrastructure.Telegram,
	spreadsheetID string) *CleaningService {
	return &CleaningService{
		sheets:        sheetsClient,
		Telegram:      telegramClient,
		spreadsheetID: spreadsheetID,
	}
}

func (s *CleaningService) StartNewWeek() {
	now := time.Now()
	sheetTitle := fmt.Sprintf("%s-%s", now.Format("02.01"), now.AddDate(0, 0, 6).Format("02.01"))
	err := s.sheets.CreateNewSheet(s.spreadsheetID, sheetTitle)
	if err != nil {
	}

}

func (s *CleaningService) HandleTaskBooking(memberID int64, taskID int) {
}

func (s *CleaningService) HandleTaskCompletion(memberID int64, taskID int) {
}

func (s *CleaningService) HandleTaskVerification(memberID int64, taskID int) {
}
