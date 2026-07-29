package scheduler

import (
	"context"
	"log"
	"time"

	"dorm/pkg/core/ports"
	"dorm/pkg/infrastructure/config"

	"github.com/robfig/cron/v3"
)

const (
	saturdayTaskReminderSpec     = "0 9 * * 6"
	sundayTakeTaskReminderSpec   = "0 9 * * 0"
	sundayFinishTaskReminderSpec = "0 18 * * 0"
	jobTimeout                   = 5 * time.Minute
)

type Scheduler struct {
	cron                *cron.Cron
	cleaningUseCase     ports.CleaningUseCase
	dutyReminderUseCase ports.DutyReminderUseCase
	cfg                 *config.AppConfig
	location            *time.Location
}

func NewScheduler(
	cleaningUseCase ports.CleaningUseCase,
	dutyReminderUseCase ports.DutyReminderUseCase,
	cfg *config.AppConfig,
) (*Scheduler, error) {
	location, err := time.LoadLocation(cfg.TZ)
	if err != nil {
		return nil, err
	}

	return &Scheduler{
		cron: cron.New(
			cron.WithLocation(location),
			cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)),
		),
		cleaningUseCase:     cleaningUseCase,
		dutyReminderUseCase: dutyReminderUseCase,
		cfg:                 cfg,
		location:            location,
	}, nil
}

func (s *Scheduler) Start() {
	if err := s.registerWeekStartJob(); err != nil {
		log.Fatalf("Failed to schedule WeekStart: %v", err)
	}

	if s.cfg.NotificationSchedulerEnabled {
		if err := s.registerReminderJobs(); err != nil {
			log.Fatalf("Failed to schedule notification reminders: %v", err)
		}
	} else {
		log.Println("Notification scheduler disabled")
	}

	log.Printf("Scheduler started with schedule: %s", s.cfg.Cron.WeekStart)
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func (s *Scheduler) registerWeekStartJob() error {
	_, err := s.cron.AddFunc(s.cfg.Cron.WeekStart, func() {
		log.Println("Scheduler: Triggering StartNewWeek...")
		ctx := context.Background()
		if err := s.cleaningUseCase.StartNewWeek(ctx); err != nil {
			log.Printf("Scheduler Error (StartNewWeek): %v", err)
		}
	})
	return err
}

func (s *Scheduler) registerReminderJobs() error {
	jobs := []struct {
		spec string
		job  cron.Job
	}{
		{spec: saturdayTaskReminderSpec, job: SaturdayTaskReminderJob{service: s.dutyReminderUseCase, location: s.location}},
		{spec: sundayTakeTaskReminderSpec, job: SundayTakeTaskReminderJob{service: s.dutyReminderUseCase, location: s.location}},
		{spec: sundayFinishTaskReminderSpec, job: SundayFinishTaskReminderJob{service: s.dutyReminderUseCase, location: s.location}},
	}

	for _, item := range jobs {
		if _, err := s.cron.AddJob(item.spec, item.job); err != nil {
			return err
		}
	}

	return nil
}

type SaturdayTaskReminderJob struct {
	service  ports.DutyReminderUseCase
	location *time.Location
}

func (j SaturdayTaskReminderJob) Run() {
	runReminderJob("saturday task reminders", j.location, j.service.SendSaturdayTaskReminders)
}

type SundayTakeTaskReminderJob struct {
	service  ports.DutyReminderUseCase
	location *time.Location
}

func (j SundayTakeTaskReminderJob) Run() {
	runReminderJob("sunday take task reminders", j.location, j.service.SendSundayTakeTaskReminders)
}

type SundayFinishTaskReminderJob struct {
	service  ports.DutyReminderUseCase
	location *time.Location
}

func (j SundayFinishTaskReminderJob) Run() {
	runReminderJob("sunday finish task reminders", j.location, j.service.SendSundayFinishTaskReminders)
}

func runReminderJob(name string, location *time.Location, handler func(ctx context.Context, at time.Time) error) {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	at := time.Now()
	if location != nil {
		at = at.In(location)
	}
	if err := handler(ctx, at); err != nil {
		log.Printf("Scheduler Error (%s): %v", name, err)
	}
}
