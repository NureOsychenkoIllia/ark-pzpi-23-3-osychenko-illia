#ifndef MODELS_H
#define MODELS_H

#include <Arduino.h>

// Типи подій пасажирів
enum EventType {
    EVENT_ENTRY,  // вхід пасажира
    EVENT_EXIT    // вихід пасажира
};

// Подія пасажира
struct PassengerEvent {
    uint32_t localId;           // локальний ідентифікатор
    EventType type;             // тип події
    unsigned long timestamp;    // мітка часу (Unix timestamp)
    float latitude;             // GPS-широта
    float longitude;            // GPS-довгота
    int passengerCountAfter;    // кількість пасажирів після події
    bool isSynced;              // чи синхронізовано з сервером
};

// Рекомендація ціни
struct PriceRecommendation {
    float basePrice;            // базова ціна
    float recommendedPrice;     // рекомендована ціна
    float occupancyRate;        // завантаженість (%)
    float demandCoeff;          // коефіцієнт попиту
    float timeCoeff;            // коефіцієнт часу
    float dayCoeff;             // коефіцієнт дня
    unsigned long calculatedAt; // час розрахунку
};

// Конфігурація рейсу (отримується з сервера)
struct TripConfig {
    int64_t tripId;             // ID рейсу
    int64_t routeId;            // ID маршруту
    int busCapacity;            // місткість автобуса
    float basePrice;            // базова ціна квитка
    bool isValid;               // чи валідна конфігурація
};

// Стан пристрою
struct DeviceState {
    int currentPassengers;      // поточна кількість пасажирів
    int totalEntries;           // загальна кількість входів
    int totalExits;             // загальна кількість виходів
    bool wifiConnected;         // стан WiFi
    bool serverAvailable;       // доступність сервера
    unsigned long lastSyncTime; // час останньої синхронізації
    unsigned long lastPriceCalc;// час останнього розрахунку ціни
    uint32_t eventCounter;      // лічильник подій
};

// Результат синхронізації
struct SyncResult {
    bool success;               // успішність
    int syncedCount;            // кількість синхронізованих подій
    int lastSyncedLocalId;      // останній синхронізований ID
    String serverTime;          // час сервера
    String errorMessage;        // повідомлення про помилку
};

#endif
