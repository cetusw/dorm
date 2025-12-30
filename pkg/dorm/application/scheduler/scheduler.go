package scheduler

import (
	"log"

	"dorm/pkg/dorm/application/model"
	"dorm/pkg/dorm/application/service"
	"dorm/pkg/dorm/application/service/report"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron            *cron.Cron
	cfg             model.Config
	cleaningService *service.CleaningService
	syncService     *service.SyncService
	reportService   *report.CleaningReportService
}

func NewScheduler(
	config model.Config,
	cleaningService *service.CleaningService,
	syncService *service.SyncService,
	reportService *report.CleaningReportService,
) *Scheduler {
	return &Scheduler{
		cron:            cron.New(),
		cfg:             config,
		cleaningService: cleaningService,
		syncService:     syncService,
		reportService:   reportService,
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

		log.Println("Starting cleaning report notification...")
		err = s.reportService.ProcessWeeklyReports()
		if err != nil {
			log.Printf("Notification job error: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("Could not add 'SyncSheets' cron job: %v", err)
	}
}

func (s *Scheduler) Start() {
	s.cron.Start()
}
