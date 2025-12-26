#ifndef API_CLIENT_H
#define API_CLIENT_H

#include "config.h"
#include "models.h"
#include "auth_manager.h"
#include <WiFi.h>
#include <HTTPClient.h>
#include <ArduinoJson.h>

class ApiClient {
private:
    String serverUrl;
    HTTPClient http;
    AuthManager* authManager;

    String buildUrl(const char* endpoint) {
        return serverUrl + String(API_BASE_PATH) + String(endpoint);
    }

    void setHeaders() {
        http.addHeader("Content-Type", "application/json");
        
        // Отримуємо актуальний токен
        String token = authManager->getAccessToken();
        if (token.length() > 0) {
            http.addHeader("Authorization", "Bearer " + token);
        }
    }

public:
    ApiClient() {
        serverUrl = String("http://") + SERVER_HOST + ":" + String(SERVER_PORT);
        authManager = nullptr;
    }

    void setAuthManager(AuthManager* auth) {
        authManager = auth;
    }

    void setServer(const char* host, int port) {
        serverUrl = String("http://") + host + ":" + String(port);
    }

    // Синхронізація подій пасажирів
    SyncResult syncEvents(int64_t tripId, PassengerEvent* events, int count) {
        SyncResult result = {false, 0, 0, "", ""};
        
        if (count == 0) {
            result.success = true;
            return result;
        }

        if (!authManager || !authManager->isAuthenticated()) {
            result.errorMessage = "Device not authenticated";
            Serial.println("[API] Device not authenticated");
            return result;
        }

        // Формування JSON запиту
        JsonDocument doc;
        doc["trip_id"] = tripId;
        JsonArray eventsArray = doc["events"].to<JsonArray>();

        for (int i = 0; i < count; i++) {
            JsonObject event = eventsArray.add<JsonObject>();
            event["local_id"] = events[i].localId;
            event["event_type"] = events[i].type == EVENT_ENTRY ? "entry" : "exit";
            event["timestamp"] = "2025-12-25T12:00:00Z"; // спрощено для симуляції
            event["latitude"] = events[i].latitude;
            event["longitude"] = events[i].longitude;
            event["passenger_count_after"] = events[i].passengerCountAfter;
        }

        String jsonBody;
        serializeJson(doc, jsonBody);

        // Відправка запиту
        http.begin(buildUrl("/iot/events"));
        setHeaders();
        
        Serial.printf("[API] POST /iot/events, %d events\n", count);
        Serial.printf("[API] Body: %s\n", jsonBody.c_str());

        int httpCode = http.POST(jsonBody);
        String response = http.getString();
        http.end();

        Serial.printf("[API] Response code: %d\n", httpCode);
        Serial.printf("[API] Response: %s\n", response.c_str());

        if (httpCode == 201) {
            // Парсимо відповідь
            JsonDocument responseDoc;
            DeserializationError error = deserializeJson(responseDoc, response);
            
            if (!error) {
                result.success = true;
                result.syncedCount = responseDoc["synced_count"];
                result.lastSyncedLocalId = responseDoc["last_synced_local_id"];
                result.serverTime = responseDoc["server_time"].as<String>();
                
                Serial.printf("[API] Synced %d events\n", result.syncedCount);
            } else {
                result.errorMessage = "Failed to parse response";
            }
        } else if (httpCode == 401) {
            result.errorMessage = "Authentication failed";
            Serial.println("[API] Authentication failed, clearing token");
            authManager->clearToken();
        } else {
            result.errorMessage = "HTTP error " + String(httpCode);
        }

        return result;
    }

    // Відправка рекомендації ціни
    bool sendPriceRecommendation(int64_t tripId, const PriceRecommendation& rec) {
        if (!authManager || !authManager->isAuthenticated()) {
            Serial.println("[API] Device not authenticated for price recommendation");
            return false;
        }

        JsonDocument doc;
        doc["trip_id"] = tripId;
        doc["base_price"] = rec.basePrice;
        doc["recommended_price"] = rec.recommendedPrice;
        doc["occupancy_rate"] = rec.occupancyRate;
        doc["demand_coefficient"] = rec.demandCoeff;
        doc["time_coefficient"] = rec.timeCoeff;
        doc["day_coefficient"] = rec.dayCoeff;

        String jsonBody;
        serializeJson(doc, jsonBody);

        http.begin(buildUrl("/iot/price"));
        setHeaders();

        Serial.printf("[API] POST /iot/price\n");
        Serial.printf("[API] Body: %s\n", jsonBody.c_str());

        int httpCode = http.POST(jsonBody);
        String response = http.getString();
        http.end();

        Serial.printf("[API] Response code: %d\n", httpCode);

        if (httpCode == 200) {
            Serial.println("[API] Price recommendation sent");
            return true;
        } else if (httpCode == 401) {
            Serial.println("[API] Authentication failed for price recommendation");
            authManager->clearToken();
            return false;
        } else {
            Serial.printf("[API] Failed to send price recommendation: %d\n", httpCode);
            return false;
        }
    }

    // Отримання конфігурації рейсу
    TripConfig getTripConfig(int64_t tripId) {
        TripConfig config;
        config.isValid = false;

        if (!authManager || !authManager->isAuthenticated()) {
            Serial.println("[API] Device not authenticated for trip config");
            return config;
        }

        String url = buildUrl("/iot/config/") + String(tripId);
        http.begin(url);
        setHeaders();

        Serial.printf("[API] GET /iot/config/%lld\n", tripId);

        int httpCode = http.GET();
        String response = http.getString();
        http.end();

        Serial.printf("[API] Response code: %d\n", httpCode);

        if (httpCode == 200) {
            JsonDocument doc;
            DeserializationError error = deserializeJson(doc, response);
            
            if (!error) {
                config.tripId = doc["trip_id"];
                config.routeId = doc["route_id"];
                config.busCapacity = doc["bus_capacity"];
                config.basePrice = doc["base_price"];
                config.isValid = true;

                Serial.printf("[API] Configuration: capacity=%d, basePrice=%.2f\n", 
                    config.busCapacity, config.basePrice);
            }
        } else if (httpCode == 401) {
            Serial.println("[API] Authentication failed for trip config");
            authManager->clearToken();
        } else {
            Serial.printf("[API] Failed to get trip config: %d\n", httpCode);
        }

        return config;
    }

    // Перевірка доступності сервера
    bool checkServerAvailability() {
        return WiFi.status() == WL_CONNECTED && authManager && authManager->isAuthenticated();
    }
};

#endif
