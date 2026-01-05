package telegram

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports"
)

func getSimplifiedProgress(ctx context.Context, useCase ports.CleaningUseCase, userID uuid.UUID) string {
	stats, err := useCase.GetUserStats(ctx, userID)
	if err != nil {
		// In case of an error, return an empty string. The error will be logged elsewhere if necessary.
		return ""
	}
	return fmt.Sprintf("Задач взято: %d/%.1f\n\n", stats.TotalPoints, stats.RequiredPoints)
}

func getFullProgress(ctx context.Context, useCase ports.CleaningUseCase, user *user.User) string {
	stats, err := useCase.GetUserStats(ctx, user.ID())
	if err != nil {
		return ""
	}

	status := "🟢 Норма выполнена"
	if !stats.IsQuotaMet() {
		status = "🔴 Норма не выполнена"
	}

	return fmt.Sprintf(
		"👤 %s %s\n\n📊 Баллы: %d\n🎯 Цель: %.1f\n%s",
		user.FirstName(), user.LastName(),
		stats.ConfirmedPoints, stats.RequiredPoints,
		status,
	)
}
