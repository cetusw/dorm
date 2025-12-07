package scheduler

import (
	"log"

	"github.com/robfig/cron/v3"

	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/application/service"
)

type Scheduler struct {
	cron            *cron.Cron
	cfg             model.Config
	cleaningService *service.CleaningService
	syncService     *service.SyncService
}

func NewScheduler(
	cleaningService *service.CleaningService,
	syncService *service.SyncService,
	config model.Config,
) *Scheduler {
	return &Scheduler{
		cron:            cron.New(),
		cleaningService: cleaningService,
		syncService:     syncService,
		cfg:             config,
	}
}

func (s *Scheduler) RegisterJobs() {
	s.startNewWeekJob()
	s.startSyncJob()
}

func (s *Scheduler) startNewWeekJob() {
	_, err := s.cron.AddFunc(s.cfg.WeekStart, func() {
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

func (s *Scheduler) startSyncJob() {
	_, err := s.cron.AddFunc(s.cfg.SyncStart, func() {
		log.Println("Cron job triggered: SyncSheets")
		err := s.syncService.SyncAllActiveDuties()
		if err != nil {
			log.Printf("Sync job error: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("Could not add 'SyncSheets' cron job: %v", err)
	}
}

func (s *Scheduler) Start() {
	s.cron.Start()
}
