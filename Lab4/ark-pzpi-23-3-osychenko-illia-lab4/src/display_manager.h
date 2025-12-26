// display_manager.h - Управління LCD дисплеєм
#ifndef DISPLAY_MANAGER_H
#define DISPLAY_MANAGER_H

#include "config.h"
#include "models.h"
#include <Wire.h>
#include <LiquidCrystal_I2C.h>

class DisplayManager {
private:
    LiquidCrystal_I2C lcd;
    unsigned long lastUpdate;
    int currentScreen;

public:
    DisplayManager() : lcd(LCD_I2C_ADDRESS, LCD_COLUMNS, LCD_ROWS), lastUpdate(0), currentScreen(0) {}

    void begin() {
        Wire.begin(PIN_LCD_SDA, PIN_LCD_SCL);
        lcd.init();
        lcd.backlight();
        showStartup();
        Serial.println("[Display] LCD initialized");
    }

    void showStartup() {
        lcd.clear();
        lcd.setCursor(0, 0);
        lcd.print("BusOptima IoT");
        lcd.setCursor(0, 1);
        lcd.print("v" FIRMWARE_VERSION);
    }

    void showConnecting() {
        lcd.clear();
        lcd.setCursor(0, 0);
        lcd.print("Connecting...");
        lcd.setCursor(0, 1);
        lcd.print(WIFI_SSID);
    }

    void showConnected(const char* ip) {
        lcd.clear();
        lcd.setCursor(0, 0);
        lcd.print("WiFi OK");
        lcd.setCursor(0, 1);
        lcd.print(ip);
    }

    void showError(const char* message) {
        lcd.clear();
        lcd.setCursor(0, 0);
        lcd.print("ERROR:");
        lcd.setCursor(0, 1);
        lcd.print(message);
    }

    // Основний екран з інформацією про пасажирів та ціну
    void showMainScreen(const DeviceState& state, const PriceRecommendation& price, int capacity, bool offlineMode = false) {
        lcd.clear();
        
        // Рядок 1: Пасажири та завантаженість
        lcd.setCursor(0, 0);
        char line1[17];
        snprintf(line1, sizeof(line1), "Pas:%d/%d %d%%", 
            state.currentPassengers, capacity, 
            (int)price.occupancyRate);
        lcd.print(line1);

        // Рядок 2: Ціна та статус роботи
        // Статуси: [O] - Offline, [A] - Authenticated, [X] - WiFi OK/No Auth, [W] - No WiFi
        lcd.setCursor(0, 1);
        char line2[17];
        char statusChar;
        if (offlineMode) {
            statusChar = 'O';  // Автономний режим
        } else if (state.wifiConnected && state.serverAvailable) {
            statusChar = 'A';  // Автентифіковано
        } else if (state.wifiConnected) {
            statusChar = 'X';  // WiFi підключено, але не автентифіковано
        } else {
            statusChar = 'W';  // Немає WiFi
        }
        snprintf(line2, sizeof(line2), "%.0f UAH [%c]", price.recommendedPrice, statusChar);
        lcd.print(line2);
    }

    // Екран синхронізації
    void showSyncScreen(int pending, bool syncing) {
        lcd.clear();
        lcd.setCursor(0, 0);
        lcd.print(syncing ? "Syncing..." : "Sync Status");
        lcd.setCursor(0, 1);
        char line[17];
        snprintf(line, sizeof(line), "Pending: %d", pending);
        lcd.print(line);
    }

    // Екран події пасажира
    void showPassengerEvent(EventType type, int count) {
        lcd.clear();
        lcd.setCursor(0, 0);
        lcd.print(type == EVENT_ENTRY ? ">> ENTRY" : "<< EXIT");
        lcd.setCursor(0, 1);
        char line[17];
        snprintf(line, sizeof(line), "Total: %d", count);
        lcd.print(line);
    }

    // Екран статистики пам'яті
    void showStorageStats(int totalEvents, int unsyncedEvents, float usagePercent) {
        lcd.clear();
        lcd.setCursor(0, 0);
        char line1[17];
        snprintf(line1, sizeof(line1), "Ev:%d/%d %.0f%%", unsyncedEvents, totalEvents, usagePercent);
        lcd.print(line1);
        
        lcd.setCursor(0, 1);
        lcd.print("Storage Stats");
    }

    // Екран скидання налаштувань
    void showReset() {
        lcd.clear();
        lcd.setCursor(0, 0);
        lcd.print("RESET");
        lcd.setCursor(0, 1);
        lcd.print("Settings cleared");
    }

    void update() {
        // Оновлення дисплея за потреби
    }

    // Екран помилки підключення
    void showConnectionError(const char* error, int code = 0) {
        lcd.clear();
        lcd.setCursor(0, 0);
        lcd.print("CONN ERROR:");
        lcd.setCursor(0, 1);
        if (code > 0) {
            char line[17];
            snprintf(line, sizeof(line), "%s %d", error, code);
            lcd.print(line);
        } else {
            lcd.print(error);
        }
    }

    // Екран статусу автентифікації
    void showAuthStatus(const char* status, const char* detail = nullptr) {
        lcd.clear();
        lcd.setCursor(0, 0);
        lcd.print("AUTH:");
        lcd.print(status);
        if (detail) {
            lcd.setCursor(0, 1);
            lcd.print(detail);
        }
    }

    // Екран налагодження
    void showDebugInfo(const char* line1, const char* line2) {
        lcd.clear();
        lcd.setCursor(0, 0);
        lcd.print(line1);
        lcd.setCursor(0, 1);
        lcd.print(line2);
    }
};

#endif
