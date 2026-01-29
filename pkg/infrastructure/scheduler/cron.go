package scheduler

import (
	"context"
	"log"

	"github.com/robfig/cron/v3"

	"dorm/pkg/core/ports"
	"dorm/pkg/infrastructure/config"
)

type Scheduler struct {
	cron            *cron.Cron
	cleaningUseCase ports.CleaningUseCase
	cfg             *config.AppConfig
}

func NewScheduler(cleaningUseCase ports.CleaningUseCase, cfg *config.AppConfig) *Scheduler {
	return &Scheduler{
		cron:            cron.New(),
		cleaningUseCase: cleaningUseCase,
		cfg:             cfg,
	}
}

func (s *Scheduler) Start() {
	_, err := s.cron.AddFunc(s.cfg.Cron.WeekStart, func() {
		log.Println("Scheduler: Triggering StartNewWeek...")
		ctx := context.Background()
		if err := s.cleaningUseCase.StartNewWeek(ctx); err != nil {
			log.Printf("Scheduler Error (StartNewWeek): %v", err)
		}
	})
	if err != nil {
		log.Fatalf("Failed to schedule WeekStart: %v", err)
	}

	log.Printf("Scheduler started with schedule: %s", s.cfg.Cron.WeekStart)
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}
