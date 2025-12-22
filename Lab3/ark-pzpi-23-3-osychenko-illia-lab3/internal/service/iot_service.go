package service

import (
	"busoptima/internal/model"
	"busoptima/internal/repository"
	"context"
	"fmt"
	"time"
)

// IoTService інтерфейс для роботи з IoT-пристроями
type IoTService interface {
	SyncEvents(ctx context.Context, tripID int64, events []model.PassengerEvent) (*SyncEventsResponse, error)
	SendPriceRecommendation(ctx context.Context, recommendation *model.PriceRecommendation) error
	GetTripConfig(ctx context.Context, tripID int64) (*TripConfig, error)
}

type SyncEventsResponse struct {
	SyncedCount           int    `json:"synced_count"`
	LastSyncedLocalID     int    `json:"last_synced_local_id"`
	TripCurrentPassengers int    `json:"trip_current_passengers"`
	ServerTime            string `json:"server_time"`
}

type TripConfig struct {
	TripID      int64   `json:"trip_id"`
	RouteID     int64   `json:"route_id"`
	BusCapacity int     `json:"bus_capacity"`
	BasePrice   float64 `json:"base_price"`
}

type iotService struct {
	deviceRepo      repository.DeviceRepository
	eventRepo       repository.PassengerEventRepository
	tripRepo        repository.TripRepository
	priceRecommRepo repository.PriceRecommendationRepository
}

func NewIoTService(deviceRepo repository.DeviceRepository, eventRepo repository.PassengerEventRepository, tripRepo repository.TripRepository, priceRecommRepo repository.PriceRecommendationRepository) IoTService {
	return &iotService{
		deviceRepo:      deviceRepo,
		eventRepo:       eventRepo,
		tripRepo:        tripRepo,
		priceRecommRepo: priceRecommRepo,
	}
}

// SyncEvents синхронізує події пасажирів від IoT-пристрою
func (s *iotService) SyncEvents(ctx context.Context, tripID int64, events []model.PassengerEvent) (*SyncEventsResponse, error) {
	if len(events) == 0 {
		return &SyncEventsResponse{
			SyncedCount:           0,
			LastSyncedLocalID:     0,
			TripCurrentPassengers: 0,
			ServerTime:            time.Now().Format(time.RFC3339),
		}, nil
	}

	// Встановлюємо trip_id для всіх подій
	for i := range events {
		events[i].TripID = tripID
	}

	// Зберігаємо події пакетом
	err := s.eventRepo.BatchCreate(ctx, events)
	if err != nil {
		return nil, fmt.Errorf("failed to sync events: %w", err)
	}

	// Отримуємо оновлену інформацію про рейс
	trip, err := s.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip: %w", err)
	}

	// Знаходимо останню подію для local_id
	lastLocalID := 0
	if len(events) > 0 && events[len(events)-1].DeviceLocalID != nil {
		lastLocalID = *events[len(events)-1].DeviceLocalID
	}

	return &SyncEventsResponse{
		SyncedCount:           len(events),
		LastSyncedLocalID:     lastLocalID,
		TripCurrentPassengers: trip.CurrentPassengers,
		ServerTime:            time.Now().Format(time.RFC3339),
	}, nil
}

// SendPriceRecommendation зберігає рекомендацію ціни від IoT-пристрою
func (s *iotService) SendPriceRecommendation(ctx context.Context, recommendation *model.PriceRecommendation) error {
	// Перевіряємо, чи існує рейс
	_, err := s.tripRepo.GetByID(ctx, recommendation.TripID)
	if err != nil {
		return fmt.Errorf("trip not found: %w", err)
	}

	// Зберігаємо рекомендацію ціни
	err = s.priceRecommRepo.Create(ctx, recommendation)
	if err != nil {
		return fmt.Errorf("failed to save price recommendation: %w", err)
	}

	return nil
}

// GetTripConfig повертає конфігурацію рейсу для IoT-пристрою
func (s *iotService) GetTripConfig(ctx context.Context, tripID int64) (*TripConfig, error) {
	trip, err := s.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip: %w", err)
	}

	config := &TripConfig{
		TripID:  tripID,
		RouteID: trip.RouteID,
	}

	if trip.Route != nil {
		config.BasePrice = trip.Route.BasePrice
	}

	if trip.Bus != nil {
		config.BusCapacity = trip.Bus.Capacity
	}

	return config, nil
}
