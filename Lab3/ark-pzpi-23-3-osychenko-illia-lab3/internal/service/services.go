package service

// Services містить всі сервіси
type Services struct {
	Auth      AuthService
	Route     RouteService
	Bus       BusService
	Trip      TripService
	IoT       IoTService
	Analytics AnalyticsService
	Forecast  ForecastService
	Pricing   PricingService
	Settings  SettingsService
}
