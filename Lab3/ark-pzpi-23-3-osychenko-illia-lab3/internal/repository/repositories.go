package repository

import "github.com/jmoiron/sqlx"

// Repositories містить всі репозиторії
type Repositories struct {
	User                UserRepository
	Route               RouteRepository
	Bus                 BusRepository
	Device              DeviceRepository
	Trip                TripRepository
	Event               PassengerEventRepository
	Analytics           AnalyticsRepository
	Audit               AuditLogRepository
	PriceRecommendation PriceRecommendationRepository
	Settings            SettingsRepository
}

// NewRepositories створює новий набір репозиторіїв
func NewRepositories(db *sqlx.DB) *Repositories {
	return &Repositories{
		User:                NewUserRepository(db),
		Route:               NewRouteRepository(db),
		Bus:                 NewBusRepository(db),
		Device:              NewDeviceRepository(db),
		Trip:                NewTripRepository(db),
		Event:               NewPassengerEventRepository(db),
		Analytics:           NewAnalyticsRepository(db),
		Audit:               NewAuditLogRepository(db),
		PriceRecommendation: NewPriceRecommendationRepository(db),
		Settings:            NewSettingsRepository(db),
	}
}
