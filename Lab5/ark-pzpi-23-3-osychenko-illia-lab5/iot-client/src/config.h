// config.h - Конфігурація IoT клієнта BusOptima
#ifndef CONFIG_H
#define CONFIG_H

// Налаштування WiFi
#define WIFI_SSID "Wokwi-GUEST"
#define WIFI_PASSWORD ""
#define WIFI_CONNECT_TIMEOUT_MS 10000
#define WIFI_RECONNECT_INTERVAL_MS 5000

// Налаштування сервера
#define SERVER_HOST "144.24.238.30"
#define SERVER_PORT 5033
#define API_BASE_PATH "/api"
#define DEVICE_SECRET_TOKEN "device_token_123"

// Режим симуляції
#define SIMULATION_MODE false  // true для тестування, false для підключення до сервера

// Налаштування пристрою
#define DEVICE_SERIAL "IOT001"  // Має відповідати зареєстрованому в БД
#define FIRMWARE_VERSION "1.0.0"

// Налаштування автентифікації
#define AUTH_TOKEN_FILE "/auth_token.dat"
#define TOKEN_REFRESH_INTERVAL_MS 3600000  // 1 година
#define TOKEN_EXPIRY_BUFFER_MS 300000      // 5 хвилин буфер до закінчення

// Налаштування рейсу (за замовчуванням)
#define DEFAULT_TRIP_ID 1
#define DEFAULT_BUS_CAPACITY 50
#define DEFAULT_BASE_PRICE 200.0

// GPIO піни
#define PIN_PIR_ENTRY 14
#define PIN_PIR_EXIT 12
#define PIN_LED_STATUS 25
#define PIN_LED_WIFI 26
#define PIN_LED_ERROR 27
#define PIN_BTN_RESET 4
#define PIN_BTN_SYNC 5
#define PIN_LCD_SDA 21
#define PIN_LCD_SCL 22

// Інтервали (мс)
#define SYNC_INTERVAL_MS 300000        // 5 хвилин
#define PRICE_CALC_INTERVAL_MS 300000  // 5 хвилин
#define DISPLAY_UPDATE_INTERVAL_MS 1000
#define SENSOR_DEBOUNCE_MS 500

// Буфер подій
#define MAX_EVENTS_BUFFER 10000        // Збільшено для тижневої автономії
#define EVENTS_BATCH_SIZE 100          // Збільшено розмір пакету
#define EVENTS_FILE_PATH "/events.dat" // Файл для зберігання подій
#define CONFIG_FILE_PATH "/config.dat" // Файл для конфігурації

// Коефіцієнти динамічного ціноутворення
#define DEMAND_COEFF_LOW 0.75      // завантаженість < 30%
#define DEMAND_COEFF_MEDIUM 0.95   // 30% <= завантаженість < 60%
#define DEMAND_COEFF_HIGH 1.10     // 60% <= завантаженість < 85%
#define DEMAND_COEFF_VERY_HIGH 1.40 // завантаженість >= 85%

#define TIME_COEFF_PEAK 1.20       // пікові години (7-9, 17-19)
#define TIME_COEFF_NIGHT 0.80      // нічні години (23-6)
#define TIME_COEFF_NORMAL 1.00     // звичайний час

#define DAY_COEFF_WEEKEND 1.15     // вихідні
#define DAY_COEFF_WEEKDAY 1.00     // будні

#define PRICE_MIN_COEFF 0.70       // мінімальна ціна (70% від базової)
#define PRICE_MAX_COEFF 1.50       // максимальна ціна (150% від базової)
#define PRICE_ROUND_STEP 5.0       // округлення до 5 грн

// Пороги завантаженості для коефіцієнтів попиту (%)
#define OCCUPANCY_LOW_THRESHOLD 30.0
#define OCCUPANCY_MEDIUM_THRESHOLD 60.0
#define OCCUPANCY_HIGH_THRESHOLD 85.0

// Пікові години
#define PEAK_MORNING_START 7
#define PEAK_MORNING_END 9
#define PEAK_EVENING_START 17
#define PEAK_EVENING_END 19
#define NIGHT_START 23
#define NIGHT_END 6

// Пороги категорій ціни
#define PRICE_CATEGORY_DISCOUNT 0.80
#define PRICE_CATEGORY_LOW 0.95
#define PRICE_CATEGORY_NORMAL 1.05
#define PRICE_CATEGORY_HIGH 1.20

// Таймери та інтервали
#define PIR_TIMEOUT_SECONDS 2
#define HEARTBEAT_INTERVAL_MS 30000
#define STATUS_DISPLAY_INTERVAL_MS 10000
#define TOKEN_CHECK_INTERVAL_MS 600000      // 10 хвилин
#define STORAGE_CHECK_INTERVAL_MS 1800000   // 30 хвилин
#define BUTTON_DOUBLE_CLICK_MS 1000

// Затримки UI (мс)
#define DELAY_DEBOUNCE_MS 50
#define DELAY_LED_FLASH_MS 200
#define DELAY_SHORT_MS 500
#define DELAY_MEDIUM_MS 1000
#define DELAY_LONG_MS 2000
#define DELAY_STATS_DISPLAY_MS 3000

// HTTP таймаути (мс)
#define HTTP_TIMEOUT_SHORT_MS 3000
#define HTTP_TIMEOUT_NORMAL_MS 10000
#define HTTP_TIMEOUT_LONG_MS 15000

// Пороги буфера
#define BUFFER_COMPACT_THRESHOLD 0.8
#define BUFFER_EMERGENCY_THRESHOLD 0.5
#define STORAGE_WARNING_PERCENT 90.0

// LCD параметри
#define LCD_I2C_ADDRESS 0x27
#define LCD_COLUMNS 16
#define LCD_ROWS 2

// Часовий пояс (секунди від UTC)
#define TIMEZONE_OFFSET_HOURS 2

// Резервний буфер у пам'яті
#define MEMORY_BUFFER_SIZE 100
#define FILE_COPY_BUFFER_SIZE 512

#endif
