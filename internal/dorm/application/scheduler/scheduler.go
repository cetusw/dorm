package scheduler

import (
	"log"

	"github.com/robfig/cron/v3"

	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/application/service"
)

type Scheduler struct {
	cron            *cron.Cron
	cleaningService *service.CleaningService
	weekStart       string
}

func NewScheduler(cleaningService *service.CleaningService, config model.Config) *Scheduler {
	return &Scheduler{
		cron:            cron.New(),
		cleaningService: cleaningService,
		weekStart:       config.WeekStart,
	}
}

func (s *Scheduler) RegisterJobs() {
	s.startNewWeekJob()
}

func (s *Scheduler) startNewWeekJob() {
	_, err := s.cron.AddFunc(s.weekStart, func() {
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
