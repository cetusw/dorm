package telegram

import (
	"context"
	"dorm/pkg/core/ports/dto"
	"fmt"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports"
)

func getCoveredProgress(ctx context.Context, useCase ports.CleaningUseCase, userID uuid.UUID) string {
	stats, err := useCase.GetUserStats(ctx, userID)
	if err != nil {
		return ""
	}
	status := "🟢"
	if !stats.IsQuotaCovered() {
		status = "🔴"
	}
	return fmt.Sprintf("%s Взято: %d/%.1f\n", status, stats.TotalPoints, stats.RequiredPoints)
}

func getUserProgress(ctx context.Context, useCase ports.CleaningUseCase, userID uuid.UUID) string {
	stats, err := useCase.GetUserStats(ctx, userID)
	if err != nil {
		return ""
	}
	status := "🟢"
	if !stats.IsUserProgressMet() {
		status = "🔴"
	}
	return fmt.Sprintf("%s Выполнено: %d/%d\n", status, stats.ConfirmedPoints, stats.TotalPoints)
}

func getMetProgress(ctx context.Context, useCase ports.CleaningUseCase, userID uuid.UUID) string {
	stats, err := useCase.GetUserStats(ctx, userID)
	if err != nil {
		return ""
	}
	status := "🟢"
	if !stats.IsQuotaMet() {
		status = "🔴"
	}
	return fmt.Sprintf("%s Выполнено: %d/%.1f\n", status, stats.ConfirmedPoints, stats.RequiredPoints)
}

func getFullProgress(ctx context.Context, useCase ports.CleaningUseCase, user *user.User) string {
	return fmt.Sprintf(
		"👤 %s %s\n\n*Прогресс:*\n%s%s\n",
		user.FirstName(),
		user.LastName(),
		getCoveredProgress(ctx, useCase, user.ID()),
		getUserProgress(ctx, useCase, user.ID()),
	)
}

func filterTasksByArea(tasks []dto.TaskViewModel, areaID int) []dto.TaskViewModel {
	var filtered []dto.TaskViewModel
	for _, t := range tasks {
		if t.AreaID == areaID {
			filtered = append(filtered, t)
		}
	}
	return filtered
}
