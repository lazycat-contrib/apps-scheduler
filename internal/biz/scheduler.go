package biz

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sync"
	"time"

	"apps-scheduler/internal/ent"
	"apps-scheduler/internal/pkg/serverchan"

	gohelper "gitee.com/linakesi/lzc-sdk/lang/go"
	"gitee.com/linakesi/lzc-sdk/lang/go/sys"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"
)

type Scheduler struct {
	useCase *UseCase
	stopCh  chan struct{}
	wg      sync.WaitGroup
	loc     *time.Location // Timezone location for all time operations
}

func NewScheduler(useCase *UseCase) *Scheduler {
	// Get timezone from TZ environment variable
	tz := os.Getenv("TZ")
	loc := time.Local // Default to local time

	if tz != "" {
		if parsedLoc, err := time.LoadLocation(tz); err == nil {
			loc = parsedLoc
			log.Info().Str("timezone", tz).Msg("Scheduler using configured timezone")
		} else {
			log.Warn().Err(err).Str("timezone", tz).Msg("Failed to load timezone, using local time")
		}
	} else {
		log.Info().Msg("Scheduler using system local time (TZ not set)")
	}

	return &Scheduler{
		useCase: useCase,
		stopCh:  make(chan struct{}),
		loc:     loc,
	}
}

func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.run()
	log.Info().Msg("Scheduler started")
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	log.Info().Msg("Scheduler stopped")
}

func (s *Scheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	// Run immediately on start
	s.checkAndExecute()

	for {
		select {
		case <-ticker.C:
			s.checkAndExecute()
		case <-s.stopCh:
			return
		}
	}
}

func (s *Scheduler) checkAndExecute() {
	ctx := context.Background()
	now := time.Now().In(s.loc) // Use configured timezone
	weekday := int(now.Weekday())
	hour := now.Hour()
	minute := now.Minute()

	log.Debug().
		Str("timezone", s.loc.String()).
		Int("weekday", weekday).
		Int("hour", hour).
		Int("minute", minute).
		Str("current_time", now.Format(time.RFC3339)).
		Msg("Checking schedules")

	schedules, err := s.useCase.GetEnabledSchedules(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get enabled schedules")
		return
	}

	for _, sch := range schedules {
		if s.shouldExecute(sch, weekday, hour, minute) {
			go s.executeSchedule(ctx, sch)
		}
	}
}

func (s *Scheduler) shouldExecute(sch *ent.Schedule, weekday, hour, minute int) bool {
	if sch.Hour != hour || sch.Minute != minute {
		return false
	}

	return slices.Contains(sch.WeekDays, weekday)
}

func (s *Scheduler) executeSchedule(ctx context.Context, sch *ent.Schedule) {
	log.Info().
		Str("schedule_id", sch.ID.String()).
		Str("app_id", sch.AppID).
		Str("operation", string(sch.Operation)).
		Msg("Executing scheduled task")

	var err error
	switch sch.Operation {
	case "resume":
		err = s.resumeApp(ctx, sch.AppID, sch.Creator)
	case "pause":
		err = s.pauseApp(ctx, sch.AppID, sch.Creator)
	default:
		log.Warn().Str("operation", string(sch.Operation)).Msg("Unknown operation")
		return
	}

	success := err == nil
	if err != nil {
		log.Error().Err(err).
			Str("schedule_id", sch.ID.String()).
			Str("app_id", sch.AppID).
			Msg("Failed to execute scheduled task")
	} else {
		log.Info().
			Str("schedule_id", sch.ID.String()).
			Str("app_id", sch.AppID).
			Msg("Scheduled task executed successfully")
	}

	// Send notification
	s.sendNotification(ctx, sch, success)
}

func (s *Scheduler) resumeApp(ctx context.Context, appID, userID string) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "x-hc-user-id", userID)

	gw, err := gohelper.NewAPIGateway(ctx)
	if err != nil {
		log.Error().Err(err).Str("app_id", appID).Msg("Failed to create API gateway")
		return err
	}
	defer gw.Close()

	// 查询应用当前状态
	resp, err := gw.PkgManager.QueryApplication(ctx, &sys.QueryApplicationRequest{
		AppidList: []string{appID},
	})
	if err != nil {
		log.Error().Err(err).Str("app_id", appID).Msg("Failed to query application status")
		return err
	}

	// 检查应用是否存在
	if len(resp.InfoList) == 0 {
		log.Warn().Str("app_id", appID).Msg("Application not found")
		return fmt.Errorf("application %s not found", appID)
	}

	appInfo := resp.InfoList[0]
	log.Info().
		Str("app_id", appID).
		Str("app_status", appInfo.Status.String()).
		Str("instance_status", appInfo.InstanceStatus.String()).
		Msg("Current application status")

	// 检查应用是否已安装
	if appInfo.Status != sys.AppStatus_Installed {
		log.Warn().
			Str("app_id", appID).
			Str("status", appInfo.Status.String()).
			Msg("Application is not installed, cannot start")
		return fmt.Errorf("application %s is not installed (status: %s)", appID, appInfo.Status.String())
	}

	// 如果应用已经在运行，直接返回成功
	if appInfo.InstanceStatus == sys.InstanceStatus_Status_Running {
		log.Info().Str("app_id", appID).Msg("Application is already running")
		return nil
	}

	// 如果应用正在恢复中，等待一下
	if appInfo.InstanceStatus == sys.InstanceStatus_Status_Starting {
		log.Info().Str("app_id", appID).Msg("Application is already starting, waiting...")
		return nil
	}

	// 调用 Resume 恢复应用
	_, err = gw.PkgManager.Resume(ctx, &sys.AppInstance{
		Appid: appID,
		Uid:   userID,
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("app_id", appID).
			Str("instance_status", appInfo.InstanceStatus.String()).
			Msg("Failed to resume application")
		return err
	}

	log.Info().Str("app_id", appID).Msg("Successfully called Resume for application")
	return nil
}

func (s *Scheduler) pauseApp(ctx context.Context, appID, userID string) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "x-hc-user-id", userID)

	gw, err := gohelper.NewAPIGateway(ctx)
	if err != nil {
		return err
	}
	defer gw.Close()

	_, err = gw.PkgManager.Pause(ctx, &sys.AppInstance{
		Appid: appID,
		Uid:   userID,
	})
	return err
}

func (s *Scheduler) sendNotification(ctx context.Context, sch *ent.Schedule, success bool) {
	notifyConfig, err := s.useCase.GetNotifyConfig(ctx, sch.Creator)
	if err != nil {
		return // No notification configured
	}

	if !notifyConfig.Enabled {
		return
	}

	if success && !notifyConfig.OnSuccess {
		return
	}
	if !success && !notifyConfig.OnFailure {
		return
	}

	client := serverchan.NewClient(notifyConfig.SendKey)
	if err := client.SendAppOperation(sch.AppTitle, string(sch.Operation), success); err != nil {
		log.Warn().Err(err).Msg("Failed to send notification")
	}
}
