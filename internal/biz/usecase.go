package biz

import (
	"context"
	"fmt"

	"apps-scheduler/internal/ent"
	"apps-scheduler/internal/ent/notifyconfig"
	"apps-scheduler/internal/ent/schedule"

	"github.com/google/uuid"
	_ "github.com/lib-x/entsqlite"
	"github.com/rs/zerolog/log"
)

type UseCase struct {
	client *ent.Client
}

func NewUseCase(dbPath string) (*UseCase, error) {
	dataSourceName := fmt.Sprintf(
		"file:%s?cache=shared&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(10000)",
		dbPath,
	)

	client, err := ent.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := client.Schema.Create(context.Background()); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	log.Info().Str("path", dbPath).Msg("Database initialized")

	return &UseCase{client: client}, nil
}

func (u *UseCase) Close() error {
	return u.client.Close()
}

func (u *UseCase) Client() *ent.Client {
	return u.client
}

// Schedule operations

func (u *UseCase) CreateSchedule(ctx context.Context, name, appID, appTitle, operation, creator string, weekDays []int, hour, minute int) (*ent.Schedule, error) {
	return u.client.Schedule.Create().
		SetName(name).
		SetAppID(appID).
		SetAppTitle(appTitle).
		SetOperation(schedule.Operation(operation)).
		SetWeekDays(weekDays).
		SetHour(hour).
		SetMinute(minute).
		SetCreator(creator).
		SetEnabled(true).
		Save(ctx)
}

func (u *UseCase) GetSchedule(ctx context.Context, id uuid.UUID) (*ent.Schedule, error) {
	return u.client.Schedule.Get(ctx, id)
}

func (u *UseCase) ListSchedules(ctx context.Context) ([]*ent.Schedule, error) {
	return u.client.Schedule.Query().
		Order(ent.Desc(schedule.FieldCreatedAt)).
		All(ctx)
}

func (u *UseCase) ListSchedulesByUser(ctx context.Context, userID string) ([]*ent.Schedule, error) {
	return u.client.Schedule.Query().
		Where(schedule.Creator(userID)).
		Order(ent.Desc(schedule.FieldCreatedAt)).
		All(ctx)
}

func (u *UseCase) UpdateSchedule(ctx context.Context, id uuid.UUID, name, appID, appTitle, operation string, weekDays []int, hour, minute int, enabled bool) (*ent.Schedule, error) {
	return u.client.Schedule.UpdateOneID(id).
		SetName(name).
		SetAppID(appID).
		SetAppTitle(appTitle).
		SetOperation(schedule.Operation(operation)).
		SetWeekDays(weekDays).
		SetHour(hour).
		SetMinute(minute).
		SetEnabled(enabled).
		Save(ctx)
}

func (u *UseCase) ToggleSchedule(ctx context.Context, id uuid.UUID, enabled bool) (*ent.Schedule, error) {
	return u.client.Schedule.UpdateOneID(id).
		SetEnabled(enabled).
		Save(ctx)
}

func (u *UseCase) DeleteSchedule(ctx context.Context, id uuid.UUID) error {
	return u.client.Schedule.DeleteOneID(id).Exec(ctx)
}

func (u *UseCase) GetEnabledSchedules(ctx context.Context) ([]*ent.Schedule, error) {
	return u.client.Schedule.Query().
		Where(schedule.Enabled(true)).
		All(ctx)
}

// NotifyConfig operations

func (u *UseCase) GetNotifyConfig(ctx context.Context, userID string) (*ent.NotifyConfig, error) {
	return u.client.NotifyConfig.Query().
		Where(notifyconfig.UserID(userID)).
		Only(ctx)
}

func (u *UseCase) SaveNotifyConfig(ctx context.Context, userID, sendKey string, enabled, onSuccess, onFailure bool) (*ent.NotifyConfig, error) {
	// Try to update existing
	existing, err := u.GetNotifyConfig(ctx, userID)
	if err == nil {
		return u.client.NotifyConfig.UpdateOne(existing).
			SetSendKey(sendKey).
			SetEnabled(enabled).
			SetOnSuccess(onSuccess).
			SetOnFailure(onFailure).
			Save(ctx)
	}

	// Create new
	return u.client.NotifyConfig.Create().
		SetUserID(userID).
		SetSendKey(sendKey).
		SetEnabled(enabled).
		SetOnSuccess(onSuccess).
		SetOnFailure(onFailure).
		Save(ctx)
}
