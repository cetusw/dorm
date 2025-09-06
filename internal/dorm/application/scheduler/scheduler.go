package scheduler

import (
	"dorm/internal/dorm/application/service"
	"log"

	"github.com/robfig/cron/v3"
)

const (
	weekStart = "20 23 * * 6"
)

type Scheduler struct {
	cron            *cron.Cron
	cleaningService *service.CleaningService
}

func NewScheduler(cleaningService *service.CleaningService) *Scheduler {
	return &Scheduler{
		cron:            cron.New(),
		cleaningService: cleaningService,
	}
}

func (s *Scheduler) RegisterJobs() {
	s.startNewWeekJob()
}

func (s *Scheduler) startNewWeekJob() {
	_, err := s.cron.AddFunc(weekStart, func() {
		log.Println("Cron job triggered: StartNewWeek")
		err := s.cleaningService.StartNewWeek()
		if err != nil {
			log.Println(err)
		}
	})
	if err != nil {
		log.Fatalf("Could not add 'StartNewWeek' cron job: %v", err)
		return
	}
	log.Println("Job registered: StartNewWeek")
}

func (s *Scheduler) Start() {
	s.cron.Start()
}
