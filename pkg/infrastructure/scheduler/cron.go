package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
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
	mu                  sync.Mutex
	started             bool
	ctx                 context.Context
	cancel              context.CancelFunc
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

	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		ctx:    ctx,
		cancel: cancel,
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

func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ctx.Err(); err != nil {
		return fmt.Errorf("start stopped scheduler: %w", err)
	}
	if s.started {
		return nil
	}
	if err := s.registerWeekStartJob(); err != nil {
		return fmt.Errorf("register week start job: %w", err)
	}

	if s.cfg.NotificationSchedulerEnabled {
		if err := s.registerReminderJobs(); err != nil {
			return fmt.Errorf("register notification reminders: %w", err)
		}
	} else {
		log.Println("Notification scheduler disabled")
	}

	log.Printf("Scheduler started with schedule: %s", s.cfg.Cron.WeekStart)
	s.cron.Start()
	s.started = true
	return nil
}

func (s *Scheduler) Stop() context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()
	done := s.cron.Stop()
	s.cancel()
	return done
}

func (s *Scheduler) registerWeekStartJob() error {
	_, err := s.cron.AddFunc(s.cfg.Cron.WeekStart, func() {
		if s.ctx.Err() != nil {
			return
		}
		log.Println("Scheduler: Triggering StartNewWeek...")
		if err := s.cleaningUseCase.StartNewWeek(s.ctx); err != nil {
			if s.ctx.Err() != nil && errors.Is(err, context.Canceled) {
				log.Println("Scheduler: StartNewWeek cancelled during shutdown")
				return
			}
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
		{spec: saturdayTaskReminderSpec, job: SaturdayTaskReminderJob{ctx: s.ctx, service: s.dutyReminderUseCase, location: s.location}},
		{spec: sundayTakeTaskReminderSpec, job: SundayTakeTaskReminderJob{ctx: s.ctx, service: s.dutyReminderUseCase, location: s.location}},
		{spec: sundayFinishTaskReminderSpec, job: SundayFinishTaskReminderJob{ctx: s.ctx, service: s.dutyReminderUseCase, location: s.location}},
	}

	for _, item := range jobs {
		if _, err := s.cron.AddJob(item.spec, item.job); err != nil {
			return err
		}
	}

	return nil
}

type SaturdayTaskReminderJob struct {
	ctx      context.Context
	service  ports.DutyReminderUseCase
	location *time.Location
}

func (j SaturdayTaskReminderJob) Run() {
	runReminderJob(j.ctx, "saturday task reminders", j.location, j.service.SendSaturdayTaskReminders)
}

type SundayTakeTaskReminderJob struct {
	ctx      context.Context
	service  ports.DutyReminderUseCase
	location *time.Location
}

func (j SundayTakeTaskReminderJob) Run() {
	runReminderJob(j.ctx, "sunday take task reminders", j.location, j.service.SendSundayTakeTaskReminders)
}

type SundayFinishTaskReminderJob struct {
	ctx      context.Context
	service  ports.DutyReminderUseCase
	location *time.Location
}

func (j SundayFinishTaskReminderJob) Run() {
	runReminderJob(j.ctx, "sunday finish task reminders", j.location, j.service.SendSundayFinishTaskReminders)
}

func runReminderJob(parent context.Context, name string, location *time.Location, handler func(ctx context.Context, at time.Time) error) {
	if parent.Err() != nil {
		return
	}
	ctx, cancel := context.WithTimeout(parent, jobTimeout)
	defer cancel()

	at := time.Now()
	if location != nil {
		at = at.In(location)
	}
	if err := handler(ctx, at); err != nil {
		if parent.Err() != nil && errors.Is(err, context.Canceled) {
			log.Printf("Scheduler: %s cancelled during shutdown", name)
			return
		}
		log.Printf("Scheduler Error (%s): %v", name, err)
	}
}
