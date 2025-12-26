#ifndef AUTH_MANAGER_H
#define AUTH_MANAGER_H

#include "config.h"
#include <LittleFS.h>
#include <WiFi.h>
#include <HTTPClient.h>
#include <ArduinoJson.h>

// Структура для зберігання токену
struct AuthToken {
    String accessToken;
    uint32_t expiresAt;     // Unix-мітка часу закінчення дії токена
    uint32_t deviceId;
    bool isValid;
};

class AuthManager {
private:
    AuthToken currentToken;
    String serverUrl;
    HTTPClient http;
    
    // Для відображення помилок
    int lastErrorCode;
    String lastErrorMessage;
    
    // Побудова URL для API
    String buildUrl(const char* endpoint) {
        return serverUrl + String(API_BASE_PATH) + String(endpoint);
    }
    
    // Завантаження токену з файлу
    bool loadTokenFromFile() {
        if (!LittleFS.exists(AUTH_TOKEN_FILE)) {
            Serial.println("[Auth] No saved token found");
            return false;
        }
        
        File tokenFile = LittleFS.open(AUTH_TOKEN_FILE, "r");
        if (!tokenFile) {
            Serial.println("[Auth] Failed to open token file");
            return false;
        }
        
        JsonDocument doc;
        DeserializationError error = deserializeJson(doc, tokenFile);
        tokenFile.close();
        
        if (error) {
            Serial.printf("[Auth] Failed to parse token file: %s\n", error.c_str());
            return false;
        }
        
        currentToken.accessToken = doc["access_token"].as<String>();
        currentToken.expiresAt = doc["expires_at"];
        currentToken.deviceId = doc["device_id"];
        currentToken.isValid = true;
        
        Serial.printf("[Auth] Token loaded, expires at: %u\n", currentToken.expiresAt);
        return true;
    }
    
    // Збереження токену у файл
    bool saveTokenToFile() {
        File tokenFile = LittleFS.open(AUTH_TOKEN_FILE, "w");
        if (!tokenFile) {
            Serial.println("[Auth] Failed to create token file");
            return false;
        }
        
        JsonDocument doc;
        doc["access_token"] = currentToken.accessToken;
        doc["expires_at"] = currentToken.expiresAt;
        doc["device_id"] = currentToken.deviceId;
        
        if (serializeJson(doc, tokenFile) == 0) {
            Serial.println("[Auth] Failed to write token file");
            tokenFile.close();
            return false;
        }
        
        tokenFile.close();
        Serial.println("[Auth] Token saved to file");
        return true;
    }
    
    // Перевірка чи токен ще дійсний
    bool isTokenExpired() {
        if (!currentToken.isValid) {
            return true;
        }
        
        uint32_t currentTime = millis() / 1000; // Поточний час у секундах
        uint32_t bufferTime = TOKEN_EXPIRY_BUFFER_MS / 1000;
        
        return (currentTime + bufferTime) >= currentToken.expiresAt;
    }

public:
    AuthManager() {
        serverUrl = String("http://") + SERVER_HOST + ":" + String(SERVER_PORT);
        currentToken.isValid = false;
        lastErrorCode = 0;
        lastErrorMessage = "";
    }
    
    void begin() {
        // Спробуємо завантажити збережений токен
        loadTokenFromFile();
        
        // Перевіримо чи токен ще дійсний
        if (isTokenExpired()) {
            Serial.println("[Auth] Saved token is expired, will authenticate on next request");
            currentToken.isValid = false;
        } else {
            Serial.println("[Auth] Valid token found");
        }
    }
    
    // Тестування POST запиту (для діагностики)
    bool testPostRequest() {
        if (WiFi.status() != WL_CONNECTED) {
            Serial.println("[Auth] WiFi not connected for POST test");
            return false;
        }
        
        Serial.println("[Auth] Testing POST request method...");
        
        // Використовуємо httpbin.org для тестування POST
        String testUrl = "http://httpbin.org/post";
        
        HTTPClient testHttp;
        testHttp.setTimeout(HTTP_TIMEOUT_NORMAL_MS);
        
        if (!testHttp.begin(testUrl)) {
            Serial.println("[Auth] Failed to begin test HTTP connection");
            return false;
        }
        
        testHttp.addHeader("Content-Type", "application/json");
        
        String testBody = "{\"test\":\"post_method\"}";
        
        Serial.printf("[Auth] Test POST to: %s\n", testUrl.c_str());
        Serial.printf("[Auth] Test body: %s\n", testBody.c_str());
        
        int httpCode = testHttp.POST(testBody);
        String response = testHttp.getString();
        testHttp.end();
        
        Serial.printf("[Auth] Test response code: %d\n", httpCode);
        Serial.printf("[Auth] Test response length: %d\n", response.length());
        
        if (httpCode == 200) {
            Serial.println("[Auth] POST method test successful");
            return true;
        } else {
            Serial.printf("[Auth] POST method test failed: %d\n", httpCode);
            return false;
        }
    }
    
    // Автентифікація пристрою
    bool authenticateDevice() {
        if (WiFi.status() != WL_CONNECTED) {
            Serial.println("[Auth] WiFi not connected");
            lastErrorMessage = "No WiFi";
            return false;
        }
        
        Serial.println("[Auth] Authenticating device...");
        Serial.printf("[Auth] Server: %s\n", serverUrl.c_str());
        
        // Спочатку протестуємо POST метод
        if (!testPostRequest()) {
            Serial.println("[Auth] POST method test failed, but continuing...");
        }
        
        String url = buildUrl("/auth/device");
        Serial.printf("[Auth] Full URL: %s\n", url.c_str());
        
        // Створюємо новий HTTP клієнт для кожного запиту
        HTTPClient authHttp;
        authHttp.setTimeout(HTTP_TIMEOUT_LONG_MS);
        
        if (!authHttp.begin(url)) {
            Serial.println("[Auth] Failed to begin HTTP connection");
            lastErrorMessage = "HTTP Begin Failed";
            lastErrorCode = -1;
            return false;
        }
        
        // Встановлюємо заголовки
        authHttp.addHeader("Content-Type", "application/json");
        authHttp.addHeader("Accept", "application/json");
        authHttp.addHeader("User-Agent", "BusOptima-IoT/1.0");
        
        // Формуємо JSON запит
        JsonDocument doc;
        doc["serial_number"] = DEVICE_SERIAL;
        doc["token"] = DEVICE_SECRET_TOKEN;
        
        String jsonBody;
        serializeJson(doc, jsonBody);
        
        Serial.printf("[Auth] POST %s\n", url.c_str());
        Serial.printf("[Auth] Headers:\n");
        Serial.printf("  Content-Type: application/json\n");
        Serial.printf("  Accept: application/json\n");
        Serial.printf("  User-Agent: BusOptima-IoT/1.0\n");
        Serial.printf("[Auth] Body: %s\n", jsonBody.c_str());
        Serial.printf("[Auth] Device Serial: %s\n", DEVICE_SERIAL);
        Serial.printf("[Auth] Device Token: %s\n", DEVICE_SECRET_TOKEN);
        Serial.printf("[Auth] JSON Body length: %d\n", jsonBody.length());
        
        // Відправляємо POST запит з явним зазначенням методу
        Serial.println("[Auth] Sending POST request...");
        Serial.printf("[Auth] Method: POST\n");
        Serial.printf("[Auth] URL: %s\n", url.c_str());
        Serial.printf("[Auth] Content-Length: %d\n", jsonBody.length());
        
        int httpCode = authHttp.POST(jsonBody);
        String response = authHttp.getString();
        
        Serial.printf("[Auth] HTTP method used: POST\n");
        Serial.printf("[Auth] Response code: %d\n", httpCode);
        Serial.printf("[Auth] Response: %s\n", response.c_str());
        
        authHttp.end();
        
        // Зберігаємо останню помилку для відображення
        lastErrorCode = httpCode;
        
        if (httpCode > 0) {
            if (httpCode == 200) {
                // Парсимо відповідь
                JsonDocument responseDoc;
                DeserializationError error = deserializeJson(responseDoc, response);
                
                if (error) {
                    Serial.printf("[Auth] Failed to parse response: %s\n", error.c_str());
                    lastErrorMessage = "Parse error";
                    return false;
                }
                
                // Зберігаємо токен
                currentToken.accessToken = responseDoc["access_token"].as<String>();
                currentToken.deviceId = responseDoc["device_id"];
                
                // Розраховуємо час закінчення токену
                uint32_t expiresIn = responseDoc["expires_in"]; // секунди
                currentToken.expiresAt = (millis() / 1000) + expiresIn;
                currentToken.isValid = true;
                
                // Зберігаємо у файл
                saveTokenToFile();
                
                Serial.printf("[Auth] Authentication successful! Device ID: %u\n", currentToken.deviceId);
                lastErrorMessage = ""; // Очищуємо помилку
                return true;
            } else {
                lastErrorMessage = "HTTP " + String(httpCode);
                Serial.printf("[Auth] Authentication failed with code: %d\n", httpCode);
                
                // Додаткова інформація для діагностики
                if (httpCode == 401) {
                    Serial.println("[Auth] 401 Unauthorized - check device credentials");
                } else if (httpCode == 404) {
                    Serial.println("[Auth] 404 Not Found - check API endpoint");
                } else if (httpCode == 405) {
                    Serial.println("[Auth] 405 Method Not Allowed - server doesn't accept POST?");
                }
                
                return false;
            }
        } else {
            // Помилка підключення
            switch (httpCode) {
                case HTTPC_ERROR_CONNECTION_REFUSED:
                    lastErrorMessage = "Connection refused";
                    Serial.println("[Auth] Connection refused - server not reachable");
                    break;
                case HTTPC_ERROR_SEND_HEADER_FAILED:
                    lastErrorMessage = "Send header failed";
                    break;
                case HTTPC_ERROR_SEND_PAYLOAD_FAILED:
                    lastErrorMessage = "Send payload failed";
                    break;
                case HTTPC_ERROR_NOT_CONNECTED:
                    lastErrorMessage = "Not connected";
                    break;
                case HTTPC_ERROR_CONNECTION_LOST:
                    lastErrorMessage = "Connection lost";
                    break;
                case HTTPC_ERROR_NO_STREAM:
                    lastErrorMessage = "No stream";
                    break;
                case HTTPC_ERROR_NO_HTTP_SERVER:
                    lastErrorMessage = "No HTTP server";
                    break;
                case HTTPC_ERROR_TOO_LESS_RAM:
                    lastErrorMessage = "Too less RAM";
                    break;
                case HTTPC_ERROR_ENCODING:
                    lastErrorMessage = "Encoding error";
                    break;
                case HTTPC_ERROR_STREAM_WRITE:
                    lastErrorMessage = "Stream write error";
                    break;
                case HTTPC_ERROR_READ_TIMEOUT:
                    lastErrorMessage = "Read timeout";
                    break;
                default:
                    lastErrorMessage = "HTTP Error " + String(httpCode);
                    break;
            }
            Serial.printf("[Auth] HTTP Error: %s (%d)\n", lastErrorMessage.c_str(), httpCode);
            return false;
        }
    }
    
    // Отримання поточного токену
    String getAccessToken() {
        // Перевіряємо чи потрібно оновити токен
        if (!currentToken.isValid || isTokenExpired()) {
            Serial.println("[Auth] Token expired or invalid, re-authenticating...");
            if (!authenticateDevice()) {
                Serial.println("[Auth] Re-authentication failed");
                return "";
            }
        }
        
        return currentToken.accessToken;
    }
    
    // Перевірка чи пристрій автентифікований
    bool isAuthenticated() {
        return currentToken.isValid && !isTokenExpired();
    }
    
    // Отримання ID пристрою
    uint32_t getDeviceId() {
        return currentToken.deviceId;
    }
    
    // Примусове оновлення токену
    bool refreshToken() {
        Serial.println("[Auth] Force token refresh...");
        currentToken.isValid = false;
        return authenticateDevice();
    }
    
    // Очищення токену (logout)
    void clearToken() {
        currentToken.isValid = false;
        currentToken.accessToken = "";
        currentToken.deviceId = 0;
        currentToken.expiresAt = 0;
        
        // Видаляємо файл токену
        if (LittleFS.exists(AUTH_TOKEN_FILE)) {
            LittleFS.remove(AUTH_TOKEN_FILE);
            Serial.println("[Auth] Token file removed");
        }
        
        Serial.println("[Auth] Token cleared");
    }
    
    // Статистика токену
    void printTokenInfo() {
        if (!currentToken.isValid) {
            Serial.println("[Auth] No valid token");
            return;
        }
        
        uint32_t currentTime = millis() / 1000;
        uint32_t timeLeft = (currentToken.expiresAt > currentTime) ? 
            (currentToken.expiresAt - currentTime) : 0;
        
        Serial.printf("[Auth] Token info:\n");
        Serial.printf("  Device ID: %u\n", currentToken.deviceId);
        Serial.printf("  Expires at: %u\n", currentToken.expiresAt);
        Serial.printf("  Time left: %u seconds\n", timeLeft);
        Serial.printf("  Is expired: %s\n", isTokenExpired() ? "Yes" : "No");
    }
    
    // Отримання інформації про останню помилку
    int getLastErrorCode() { return lastErrorCode; }
    String getLastErrorMessage() { return lastErrorMessage; }
    
    // Перевірка стану підключення
    String getConnectionStatus() {
        if (WiFi.status() != WL_CONNECTED) {
            return "No WiFi";
        }
        if (!currentToken.isValid) {
            return "No Token";
        }
        if (isTokenExpired()) {
            return "Expired";
        }
        if (!isServerOnline()) {
            return "Server Offline";
        }
        return "Connected";
    }
    
    // Детальна діагностика підключення
    String getDetailedStatus() {
        if (WiFi.status() != WL_CONNECTED) {
            return "WiFi: Disconnected";
        }
        
        if (lastErrorMessage.length() > 0) {
            return lastErrorMessage;
        }
        
        if (!currentToken.isValid) {
            return "No valid token";
        }
        
        if (isTokenExpired()) {
            return "Token expired";
        }
        
        if (!isServerOnline()) {
            return "Server offline";
        }
        
        return "All OK";
    }
    
    // Перевірка доступності сервера
    bool testServerConnection() {
        if (WiFi.status() != WL_CONNECTED) {
            Serial.println("[Auth] WiFi not connected for server test");
            lastErrorMessage = "No WiFi";
            return false;
        }
        
        Serial.println("[Auth] Testing server connection...");
        Serial.printf("[Auth] Server: %s\n", serverUrl.c_str());
        
        // Просто спробуємо підключитися до сервера без відправки запиту
        // Це дозволить перевірити доступність без створення зайвих логів
        WiFiClient client;
        
        Serial.printf("[Auth] Testing connection to %s:%d\n", SERVER_HOST, SERVER_PORT);
        
        if (client.connect(SERVER_HOST, SERVER_PORT)) {
            client.stop();
            Serial.println("[Auth] Server is reachable");
            lastErrorMessage = "";
            return true;
        } else {
            Serial.println("[Auth] Cannot connect to server");
            lastErrorMessage = "Cannot connect";
            return false;
        }
    }
    
    // Heartbeat - швидка перевірка доступності сервера через TCP
    bool heartbeat() {
        if (WiFi.status() != WL_CONNECTED) {
            lastErrorMessage = "No WiFi";
            return false;
        }
        
        WiFiClient client;
        client.setTimeout(HTTP_TIMEOUT_SHORT_MS);
        
        bool connected = client.connect(SERVER_HOST, SERVER_PORT);
        client.stop();
        
        if (connected) {
            if (lastErrorMessage == "Server offline") {
                Serial.println("[Heartbeat] Server is back online!");
                lastErrorMessage = "";
            }
            return true;
        } else {
            if (lastErrorMessage != "Server offline") {
                Serial.println("[Heartbeat] Server went offline!");
            }
            lastErrorMessage = "Server offline";
            return false;
        }
    }
    
    // Перевірка чи сервер доступний (без нового запиту)
    bool isServerOnline() {
        return lastErrorMessage != "Server offline" && 
               lastErrorMessage != "Cannot connect" &&
               lastErrorMessage != "No WiFi";
    }
};

#endif