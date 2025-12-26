#include <Arduino.h>
#include <WiFi.h>
#include <time.h>
#include <LittleFS.h>
#include "config.h"
#include "models.h"
#include "auth_manager.h"
#include "event_buffer.h"
#include "pricing_engine.h"
#include "api_client.h"
#include "display_manager.h"

// Глобальні об'єкти
AuthManager authManager;
EventBuffer eventBuffer;
PricingEngine pricingEngine;
ApiClient apiClient;
DisplayManager display;

DeviceState deviceState;
TripConfig tripConfig;
PriceRecommendation currentPrice;

// Режим роботи
bool offlineMode = false;  // Прапорець автономного режиму

// Попередні оголошення
void syncEvents();
void calculateAndSendPrice();

// Таймери
unsigned long lastSyncTime = 0;
unsigned long lastPriceCalcTime = 0;
unsigned long lastDisplayUpdate = 0;
unsigned long lastStorageCheck = 0;
unsigned long lastTokenCheck = 0;
unsigned long lastHeartbeat = 0;  // Перевірка доступності сервера

// Таймер: допоміжні змінні
unsigned long now = millis();
unsigned long lastEntryTrigger = 0;
unsigned long lastExitTrigger = 0;
boolean startEntryTimer = false;
boolean startExitTimer = false;
boolean entryMotion = false;
boolean exitMotion = false;

// Перевірка виявлення руху на датчику входу
void IRAM_ATTR detectsEntryMovement() {
    digitalWrite(PIN_LED_STATUS, HIGH);
    startEntryTimer = true;
    lastEntryTrigger = millis();
}

// Перевірка виявлення руху на датчику виходу
void IRAM_ATTR detectsExitMovement() {
    digitalWrite(PIN_LED_ERROR, HIGH);
    startExitTimer = true;
    lastExitTrigger = millis();
}

// Перевірка та перемикання в автономний режим
void checkOfflineMode() {
    bool shouldBeOffline = !deviceState.wifiConnected || 
                           !authManager.isAuthenticated() || 
                           !authManager.isServerOnline();
    
    if (shouldBeOffline && !offlineMode) {
        // Перемикаємося в автономний режим
        offlineMode = true;
        Serial.println("[System] Switching to OFFLINE mode");
        Serial.printf("[System] Events in buffer: %d (unsynced: %d)\n", 
            eventBuffer.getCount(), eventBuffer.getUnsyncedCount());
        
        // Показуємо повідомлення про автономний режим на 2 секунди
        display.showDebugInfo("OFFLINE MODE", "Events stored locally");
        delay(2000);
        
    } else if (!shouldBeOffline && offlineMode) {
        // Повертаємося в онлайн режим
        offlineMode = false;
        Serial.println("[System] Switching to ONLINE mode");
        
        // Показуємо повідомлення про онлайн режим
        display.showDebugInfo("ONLINE MODE", "Loading config...");
        delay(1000);
        
        // Підтягуємо конфігурацію з сервера
        TripConfig serverConfig = apiClient.getTripConfig(DEFAULT_TRIP_ID);
        if (serverConfig.isValid) {
            tripConfig = serverConfig;
            Serial.printf("[System] Config loaded: capacity=%d, basePrice=%.2f\n", 
                tripConfig.busCapacity, tripConfig.basePrice);
            
            // Перераховуємо ціну з новою конфігурацією
            calculateAndSendPrice();
        } else {
            Serial.println("[System] Failed to load config, using cached values");
        }
        
        // Синхронізуємо накопичені події
        display.showDebugInfo("ONLINE MODE", "Syncing events...");
        delay(1000);
        syncEvents();
    }
}

// Ініціалізація WiFi
bool initWiFi() {
    Serial.println("[WiFi] Connecting to " WIFI_SSID);
    display.showConnecting();
    
    WiFi.mode(WIFI_STA);
    WiFi.begin(WIFI_SSID, WIFI_PASSWORD);

    unsigned long startTime = millis();
    while (WiFi.status() != WL_CONNECTED) {
        if (millis() - startTime > WIFI_CONNECT_TIMEOUT_MS) {
            Serial.println("[WiFi] Connection timeout");
            return false;
        }
        digitalWrite(PIN_LED_WIFI, !digitalRead(PIN_LED_WIFI));
        delay(500);
    }

    digitalWrite(PIN_LED_WIFI, HIGH);
    Serial.printf("[WiFi] Connected, IP: %s\n", WiFi.localIP().toString().c_str());
    display.showConnected(WiFi.localIP().toString().c_str());
    
    return true;
}

// Ініціалізація часу через NTP
void initTime() {
    configTime(TIMEZONE_OFFSET_HOURS * 3600, 0, "pool.ntp.org", "time.nist.gov");
    Serial.println("[Time] Synchronizing time...");
    
    struct tm timeinfo;
    if (getLocalTime(&timeinfo, 5000)) {
        Serial.printf("[Time] Time: %02d:%02d:%02d\n", 
            timeinfo.tm_hour, timeinfo.tm_min, timeinfo.tm_sec);
    }
}

// Ініціалізація GPIO
void initGPIO() {
    // Режим датчика руху PIR: INPUT_PULLUP
    pinMode(PIN_PIR_ENTRY, INPUT_PULLUP);
    pinMode(PIN_PIR_EXIT, INPUT_PULLUP);
    
    // Налаштування пінів PIR як переривань, призначення функції переривання та режиму RISING
    attachInterrupt(digitalPinToInterrupt(PIN_PIR_ENTRY), detectsEntryMovement, RISING);
    attachInterrupt(digitalPinToInterrupt(PIN_PIR_EXIT), detectsExitMovement, RISING);
    
    // Світлодіоди
    pinMode(PIN_LED_STATUS, OUTPUT);
    pinMode(PIN_LED_WIFI, OUTPUT);
    pinMode(PIN_LED_ERROR, OUTPUT);
    
    // Встановлення світлодіодів у стан LOW
    digitalWrite(PIN_LED_STATUS, LOW);
    digitalWrite(PIN_LED_WIFI, LOW);
    digitalWrite(PIN_LED_ERROR, LOW);
    
    // Кнопки з підтяжкою
    pinMode(PIN_BTN_RESET, INPUT_PULLUP);
    pinMode(PIN_BTN_SYNC, INPUT_PULLUP);

    Serial.println("[GPIO] Pins initialized");
    Serial.printf("[GPIO] PIR Entry pin %d, PIR Exit pin %d\n", PIN_PIR_ENTRY, PIN_PIR_EXIT);
}

// Ініціалізація стану пристрою
void initDeviceState() {
    deviceState.currentPassengers = 0;
    deviceState.totalEntries = 0;
    deviceState.totalExits = 0;
    deviceState.wifiConnected = false;
    deviceState.serverAvailable = false;
    deviceState.lastSyncTime = 0;
    deviceState.lastPriceCalc = 0;
    deviceState.eventCounter = 0;

    // Завантаження конфігурації рейсу
    tripConfig.tripId = DEFAULT_TRIP_ID;
    tripConfig.busCapacity = DEFAULT_BUS_CAPACITY;
    tripConfig.basePrice = DEFAULT_BASE_PRICE;
    tripConfig.isValid = true;

    Serial.println("[State] Device state initialized");
}

// Обробка події входу пасажира
void handlePassengerEntry() {
    deviceState.currentPassengers++;
    deviceState.totalEntries++;
    
    eventBuffer.addEvent(EVENT_ENTRY, deviceState.currentPassengers);
    
    display.showPassengerEvent(EVENT_ENTRY, deviceState.currentPassengers);
    digitalWrite(PIN_LED_STATUS, HIGH);
    
    Serial.printf("[Passenger] ENTRY: current count = %d\n", deviceState.currentPassengers);
    
    delay(200);
    digitalWrite(PIN_LED_STATUS, LOW);
}

// Обробка події виходу пасажира
void handlePassengerExit() {
    if (deviceState.currentPassengers > 0) {
        deviceState.currentPassengers--;
    }
    deviceState.totalExits++;
    
    eventBuffer.addEvent(EVENT_EXIT, deviceState.currentPassengers);
    
    display.showPassengerEvent(EVENT_EXIT, deviceState.currentPassengers);
    digitalWrite(PIN_LED_STATUS, HIGH);
    
    Serial.printf("[Passenger] EXIT: current count = %d\n", deviceState.currentPassengers);
    
    delay(200);
    digitalWrite(PIN_LED_STATUS, LOW);
}

// Синхронізація подій з сервером
void syncEvents() {
    int unsyncedCount = eventBuffer.getUnsyncedCount();
    if (unsyncedCount == 0) {
        Serial.println("[Sync] No events to sync");
        return;
    }

    // Перевіряємо чи можемо синхронізувати
    if (!deviceState.wifiConnected || !authManager.isAuthenticated()) {
        Serial.printf("[Sync] Cannot sync in offline mode, %d events pending\n", unsyncedCount);
        return;
    }

    display.showSyncScreen(unsyncedCount, true);
    Serial.printf("[Sync] Syncing %d events...\n", unsyncedCount);

    PassengerEvent events[EVENTS_BATCH_SIZE];
    int count = eventBuffer.getUnsyncedEvents(events, EVENTS_BATCH_SIZE);

    SyncResult result = apiClient.syncEvents(tripConfig.tripId, events, count);

    if (result.success) {
        eventBuffer.markSynced(result.lastSyncedLocalId);
        deviceState.lastSyncTime = millis();
        digitalWrite(PIN_LED_ERROR, LOW);
        Serial.printf("[Sync] Successfully synced %d events\n", result.syncedCount);
        
        // Автоматичне компактування при великій кількості синхронізованих подій
        if (eventBuffer.getCount() > MAX_EVENTS_BUFFER * BUFFER_COMPACT_THRESHOLD) {
            Serial.println("[Sync] Buffer is getting full, compacting...");
            eventBuffer.compactBuffer();
        }
    } else {
        digitalWrite(PIN_LED_ERROR, HIGH);
        Serial.printf("[Sync] Error: %s\n", result.errorMessage.c_str());
    }
}

// Розрахунок та відправка рекомендації ціни
void calculateAndSendPrice() {
    struct tm timeinfo;
    getLocalTime(&timeinfo);

    currentPrice = pricingEngine.calculatePrice(
        tripConfig.basePrice,
        deviceState.currentPassengers,
        tripConfig.busCapacity,
        &timeinfo
    );

    deviceState.lastPriceCalc = millis();

    // Відправка на сервер (тільки якщо є підключення)
    if (deviceState.wifiConnected && authManager.isAuthenticated()) {
        if (apiClient.sendPriceRecommendation(tripConfig.tripId, currentPrice)) {
            Serial.println("[Price] Recommendation sent to server");
        } else {
            Serial.println("[Price] Failed to send recommendation, working offline");
        }
    } else {
        Serial.println("[Price] Working in offline mode, price calculated locally");
    }

    const char* category = pricingEngine.getPriceCategory(
        tripConfig.basePrice, currentPrice.recommendedPrice);
    Serial.printf("[Price] Category: %s\n", category);
}

// Обробка кнопки скидання
void handleResetButton() {
    static unsigned long lastPress = 0;
    static int pressCount = 0;
    
    if (digitalRead(PIN_BTN_RESET) == LOW) {
        delay(50); // антибрязк
        if (digitalRead(PIN_BTN_RESET) == LOW) {
            unsigned long now = millis();
            
            // Подвійне натискання для перемикання режиму
            if (now - lastPress < BUTTON_DOUBLE_CLICK_MS) {
                pressCount++;
                if (pressCount >= 2) {
                    Serial.println("[Button] Switching to simulation mode...");
                    display.showDebugInfo("Switching to", "Simulation mode");
                    
                    // Тут можна додати логіку перемикання режиму
                    // Поки що просто показуємо повідомлення
                    delay(2000);
                    pressCount = 0;
                }
            } else {
                pressCount = 1;
                Serial.println("[Button] Resetting settings and clearing auth...");
                display.showReset();
                
                // Очищуємо всі дані
                eventBuffer.clear();
                authManager.clearToken();
                initDeviceState();
                
                // Спробуємо повторно автентифікуватися
                if (deviceState.wifiConnected) {
                    if (authManager.authenticateDevice()) {
                        Serial.println("[Reset] Re-authentication successful");
                    } else {
                        Serial.println("[Reset] Re-authentication failed");
                    }
                }
                
                delay(2000);
            }
            
            lastPress = now;
        }
    }
}

// Обробка кнопки примусової синхронізації
void handleSyncButton() {
    if (digitalRead(PIN_BTN_SYNC) == LOW) {
        delay(50); // антибрязк
        if (digitalRead(PIN_BTN_SYNC) == LOW) {
            Serial.println("[Button] Force sync and auth info...");
            
            // Показуємо інформацію про підключення
            String connStatus = authManager.getConnectionStatus();
            String serverInfo = String(SERVER_HOST) + ":" + String(SERVER_PORT);
            
            display.showDebugInfo(connStatus.c_str(), serverInfo.c_str());
            delay(2000);
            
            // Показуємо інформацію про токен
            authManager.printTokenInfo();
            
            // Показуємо статистику пам'яті
            eventBuffer.printStorageStats();
            
            // Відображаємо статистику на LCD
            float usagePercent = 0.0;
            if (LittleFS.totalBytes() > 0) {
                size_t totalBytes = LittleFS.totalBytes();
                size_t usedBytes = LittleFS.usedBytes();
                usagePercent = (float)usedBytes / totalBytes * 100.0;
            } else {
                // Резервний варіант для симулятора - показуємо використання буфера
                usagePercent = (float)eventBuffer.getCount() / 100.0 * 100.0;
            }
            
            display.showStorageStats(
                eventBuffer.getCount(), 
                eventBuffer.getUnsyncedCount(), 
                usagePercent
            );
            
            delay(3000); // Показуємо статистику 3 секунди
            
            // Виконуємо синхронізацію
            if (authManager.isAuthenticated()) {
                syncEvents();
            } else {
                Serial.println("[Sync] Device not authenticated, attempting re-auth...");
                if (authManager.authenticateDevice()) {
                    syncEvents();
                } else {
                    Serial.println("[Sync] Authentication failed");
                }
            }
            
            delay(500);
        }
    }
}

// Оновлення стану WiFi
void updateWiFiStatus() {
    bool connected = WiFi.status() == WL_CONNECTED;
    
    if (connected != deviceState.wifiConnected) {
        deviceState.wifiConnected = connected;
        digitalWrite(PIN_LED_WIFI, connected ? HIGH : LOW);
        
        if (!connected) {
            Serial.println("[WiFi] Connection lost, reconnecting...");
            WiFi.reconnect();
        }
    }
}

// Налаштування
void setup() {
    Serial.begin(115200);
    delay(1000);
    
    Serial.println("\n========================================");
    Serial.println("BusOptima IoT Client v" FIRMWARE_VERSION);
    Serial.println("Device: " DEVICE_SERIAL);
    Serial.println("========================================\n");

    // Ініціалізація компонентів
    initGPIO();
    
    // Ініціалізація файлової системи
    if (!LittleFS.begin(true)) {
        Serial.println("[LittleFS] Failed to initialize, using fallback mode");
    } else {
        Serial.println("[LittleFS] Initialized successfully");
    }
    
    display.begin();
    eventBuffer.begin();
    authManager.begin();
    initDeviceState();

    // Налаштування API клієнта
    apiClient.setAuthManager(&authManager);

    // Підключення до WiFi
    if (initWiFi()) {
        deviceState.wifiConnected = true;
        initTime();
        
        // Тестування підключення до сервера
        display.showAuthStatus("Testing server...");
        if (authManager.testServerConnection()) {
            Serial.println("[Setup] Server is reachable");
            display.showAuthStatus("Server OK");
            delay(1000);
            
            // Автентифікація пристрою
            display.showAuthStatus("Connecting...");
            if (authManager.authenticateDevice()) {
                Serial.println("[Setup] Device authenticated successfully");
                display.showAuthStatus("Success!");
                delay(2000);
                
                // Завантаження конфігурації з сервера
                TripConfig serverConfig = apiClient.getTripConfig(DEFAULT_TRIP_ID);
                if (serverConfig.isValid) {
                    tripConfig = serverConfig;
                    Serial.println("[Setup] Trip configuration loaded from server");
                } else {
                    Serial.println("[Setup] Using default trip configuration");
                }
            } else {
                Serial.println("[Setup] Device authentication failed");
                display.showConnectionError("Auth Failed", authManager.getLastErrorCode());
                delay(3000);
            }
        } else {
            Serial.println("[Setup] Server is not reachable");
            display.showConnectionError("Server unreachable");
            delay(3000);
        }
    } else {
        Serial.println("[Setup] WiFi connection failed, running in offline mode");
        display.showConnectionError("No WiFi");
        delay(3000);
    }

    // Початковий розрахунок ціни
    calculateAndSendPrice();

    Serial.println("\n[System] Initialization complete");
    Serial.println("[System] Waiting for sensor events...\n");
}

// Основний цикл
void loop() {
    // Поточний час
    now = millis();

    // Перевірка датчика входу
    if((digitalRead(PIN_LED_STATUS) == HIGH) && (entryMotion == false)) {
        Serial.println("ENTRY MOTION DETECTED!!!");
        entryMotion = true;
        handlePassengerEntry();
    }

    // Перевірка датчика виходу
    if((digitalRead(PIN_LED_ERROR) == HIGH) && (exitMotion == false)) {
        Serial.println("EXIT MOTION DETECTED!!!");
        exitMotion = true;
        handlePassengerExit();
    }

    // Вимкнення світлодіода входу після таймауту PIR
    if(startEntryTimer && (now - lastEntryTrigger > (PIR_TIMEOUT_SECONDS * 1000))) {
        Serial.println("Entry motion stopped...");
        digitalWrite(PIN_LED_STATUS, LOW);
        startEntryTimer = false;
        entryMotion = false;
    }

    // Вимкнення світлодіода виходу після таймауту PIR
    if(startExitTimer && (now - lastExitTrigger > (PIR_TIMEOUT_SECONDS * 1000))) {
        Serial.println("Exit motion stopped...");
        digitalWrite(PIN_LED_ERROR, LOW);
        startExitTimer = false;
        exitMotion = false;
    }

    // Обробка кнопок
    handleResetButton();
    handleSyncButton();

    // Перевірка та перемикання режиму роботи
    checkOfflineMode();

    // Періодична синхронізація (тільки якщо є підключення)
    if (now - lastSyncTime >= SYNC_INTERVAL_MS) {
        lastSyncTime = now;
        if (deviceState.wifiConnected && authManager.isAuthenticated()) {
            syncEvents();
        } else {
            Serial.println("[Sync] Working in offline mode, events stored locally");
        }
    }

    // Періодичний розрахунок ціни
    if (now - lastPriceCalcTime >= PRICE_CALC_INTERVAL_MS) {
        lastPriceCalcTime = now;
        calculateAndSendPrice();
    }

    // Heartbeat - перевірка доступності сервера (кожні 30 секунд)
    if (now - lastHeartbeat >= HEARTBEAT_INTERVAL_MS) {
        lastHeartbeat = now;
        
        if (deviceState.wifiConnected) {
            bool serverOnline = authManager.heartbeat();
            
            if (!serverOnline && !offlineMode) {
                // Сервер став недоступним - переходимо в офлайн
                Serial.println("[Heartbeat] Server offline, switching to offline mode");
            } else if (serverOnline && offlineMode) {
                // Сервер знову доступний - спробуємо повернутися онлайн
                Serial.println("[Heartbeat] Server back online, attempting to reconnect");
                if (authManager.isAuthenticated() || authManager.authenticateDevice()) {
                    Serial.println("[Heartbeat] Reconnected successfully");
                }
            }
        }
    }

    // Оновлення дисплея
    if (now - lastDisplayUpdate >= DISPLAY_UPDATE_INTERVAL_MS) {
        lastDisplayUpdate = now;
        
        // Завжди показуємо основний екран з даними
        display.showMainScreen(deviceState, currentPrice, tripConfig.busCapacity, offlineMode);
        
        // Періодично показуємо статус підключення якщо є проблеми
        static unsigned long lastStatusShow = 0;
        if (now - lastStatusShow >= STATUS_DISPLAY_INTERVAL_MS) {
            lastStatusShow = now;
            
            String connStatus = authManager.getConnectionStatus();
            if (connStatus != "Connected") {
                String detailedStatus = authManager.getDetailedStatus();
                display.showDebugInfo(connStatus.c_str(), detailedStatus.c_str());
                delay(2000); // показуємо статус 2 секунди
                display.showMainScreen(deviceState, currentPrice, tripConfig.busCapacity, offlineMode);
            }
        }
    }

    // Періодична перевірка токену
    if (now - lastTokenCheck >= TOKEN_CHECK_INTERVAL_MS) {
        lastTokenCheck = now;
        
        if (deviceState.wifiConnected) {
            if (!authManager.isAuthenticated()) {
                Serial.println("[Token] Token expired, re-authenticating...");
                display.showAuthStatus("Re-auth...");
                
                if (authManager.authenticateDevice()) {
                    Serial.println("[Token] Re-authentication successful");
                    display.showAuthStatus("Success!");
                    delay(1000);
                } else {
                    Serial.println("[Token] Re-authentication failed");
                    display.showConnectionError("Re-auth fail", authManager.getLastErrorCode());
                    digitalWrite(PIN_LED_ERROR, HIGH);
                    delay(2000);
                }
            } else {
                Serial.println("[Token] Token is still valid");
            }
        }
    }

    // Періодична перевірка стану пам'яті
    if (now - lastStorageCheck >= STORAGE_CHECK_INTERVAL_MS) {
        lastStorageCheck = now;
        
        size_t totalBytes = LittleFS.totalBytes();
        size_t usedBytes = LittleFS.usedBytes();
        float usagePercent = (float)usedBytes / totalBytes * 100.0;
        
        Serial.printf("[Storage] Usage: %.1f%% (%u/%u bytes)\n", 
            usagePercent, usedBytes, totalBytes);
        
        // Попередження при заповненні більше порогу
        if (usagePercent > STORAGE_WARNING_PERCENT) {
            Serial.println("[Storage] WARNING: Storage almost full!");
            digitalWrite(PIN_LED_ERROR, HIGH);
            
            // Примусове компактування
            if (eventBuffer.getUnsyncedCount() < eventBuffer.getCount() * BUFFER_EMERGENCY_THRESHOLD) {
                Serial.println("[Storage] Performing emergency compaction...");
                eventBuffer.compactBuffer();
            }
        }
    }

    // Оновлення стану WiFi та автентифікації
    updateWiFiStatus();
    
    // Оновлюємо статус сервера
    deviceState.serverAvailable = authManager.isAuthenticated();

    // Індикація WiFi
    static unsigned long lastBlink = 0;
    if (now - lastBlink >= 1000) {
        lastBlink = now;
        digitalWrite(PIN_LED_WIFI, !digitalRead(PIN_LED_WIFI));
    }

    delay(10);
}
