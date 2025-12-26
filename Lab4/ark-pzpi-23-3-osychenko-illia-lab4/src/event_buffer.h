#ifndef EVENT_BUFFER_H
#define EVENT_BUFFER_H

#include "config.h"
#include "models.h"
#include <LittleFS.h>
#include <ArduinoJson.h>

// Структура для зберігання метаданих буферу
struct BufferMetadata {
    uint32_t nextLocalId;
    uint32_t totalEvents;
    uint32_t syncedEvents;
    uint32_t fileVersion;
};

class EventBuffer {
private:
    BufferMetadata metadata;
    String eventsFilePath;
    String configFilePath;
    bool littleFSAvailable;
    
    // Резервний буфер у пам'яті для симулятора
    PassengerEvent memoryBuffer[MEMORY_BUFFER_SIZE];
    int memoryCount;
    int memoryNextId;
    
    // Завантаження метаданих з файлу
    bool loadMetadata() {
        if (!LittleFS.exists(configFilePath)) {
            metadata.nextLocalId = 1;
            metadata.totalEvents = 0;
            metadata.syncedEvents = 0;
            metadata.fileVersion = 1;
            return saveMetadata();
        }
        
        File configFile = LittleFS.open(configFilePath, "r");
        if (!configFile) {
            return false;
        }
        
        size_t bytesRead = configFile.readBytes((char*)&metadata, sizeof(BufferMetadata));
        configFile.close();
        
        return bytesRead == sizeof(BufferMetadata);
    }
    
    // Збереження метаданих у файл
    bool saveMetadata() {
        File configFile = LittleFS.open(configFilePath, "w");
        if (!configFile) {
            return false;
        }
        
        size_t bytesWritten = configFile.write((uint8_t*)&metadata, sizeof(BufferMetadata));
        configFile.close();
        
        return bytesWritten == sizeof(BufferMetadata);
    }
    
    // Додавання події до файлу
    bool appendEventToFile(const PassengerEvent& event) {
        File eventsFile = LittleFS.open(eventsFilePath, "a");
        if (!eventsFile) {
            return false;
        }
        
        size_t bytesWritten = eventsFile.write((uint8_t*)&event, sizeof(PassengerEvent));
        eventsFile.close();
        
        return bytesWritten == sizeof(PassengerEvent);
    }

public:
    EventBuffer() {
        eventsFilePath = String(EVENTS_FILE_PATH);
        configFilePath = String(CONFIG_FILE_PATH);
        littleFSAvailable = false;
        memoryCount = 0;
        memoryNextId = 1;
    }

    void begin() {
        // Перевіряємо чи LittleFS доступна (true = форматувати якщо потрібно)
        littleFSAvailable = LittleFS.begin(true);
        
        if (littleFSAvailable) {
            Serial.println("[EventBuffer] LittleFS initialized successfully");
            Serial.printf("[EventBuffer] Total: %u bytes, Used: %u bytes\n", 
                LittleFS.totalBytes(), LittleFS.usedBytes());
            
            if (!loadMetadata()) {
                Serial.println("[EventBuffer] Creating new metadata");
                metadata.nextLocalId = 1;
                metadata.totalEvents = 0;
                metadata.syncedEvents = 0;
                metadata.fileVersion = 1;
                saveMetadata();
            }
            
            Serial.printf("[EventBuffer] Initialized - nextId=%u, count=%u\n", 
                metadata.nextLocalId, metadata.totalEvents);
        } else {
            Serial.println("[EventBuffer] LittleFS failed, using memory fallback");
            memoryCount = 0;
            memoryNextId = 1;
        }
    }

    // Додавання події до буферу
    bool addEvent(EventType type, int passengerCountAfter, float lat = 0.0, float lon = 0.0) {
        if (littleFSAvailable) {
            // Режим LittleFS
            if (metadata.totalEvents >= MAX_EVENTS_BUFFER) {
                if (!compactBuffer()) {
                    return false;
                }
            }

            PassengerEvent event;
            event.localId = metadata.nextLocalId++;
            event.type = type;
            event.timestamp = millis() / 1000;
            event.latitude = lat;
            event.longitude = lon;
            event.passengerCountAfter = passengerCountAfter;
            event.isSynced = false;

            if (!appendEventToFile(event)) {
                return false;
            }

            metadata.totalEvents++;
            saveMetadata();
            
            Serial.printf("[EventBuffer] Added event #%u, type=%s, passengers=%d\n",
                event.localId, type == EVENT_ENTRY ? "ENTRY" : "EXIT", passengerCountAfter);
        } else {
            // Режим резервного буфера у пам'яті
            if (memoryCount >= MEMORY_BUFFER_SIZE) {
                for (int i = 0; i < memoryCount - 1; i++) {
                    memoryBuffer[i] = memoryBuffer[i + 1];
                }
                memoryCount--;
            }
            
            PassengerEvent& event = memoryBuffer[memoryCount];
            event.localId = memoryNextId++;
            event.type = type;
            event.timestamp = millis() / 1000;
            event.latitude = lat;
            event.longitude = lon;
            event.passengerCountAfter = passengerCountAfter;
            event.isSynced = false;
            
            memoryCount++;
            
            Serial.printf("[EventBuffer] Added event #%d (memory), type=%s, passengers=%d\n",
                memoryNextId - 1, type == EVENT_ENTRY ? "ENTRY" : "EXIT", passengerCountAfter);
        }

        return true;
    }

    // Отримання несинхронізованих подій
    int getUnsyncedEvents(PassengerEvent* buffer, int maxCount) {
        if (littleFSAvailable) {
            if (!LittleFS.exists(eventsFilePath)) {
                return 0;
            }
            
            File eventsFile = LittleFS.open(eventsFilePath, "r");
            if (!eventsFile) {
                return 0;
            }
            
            int fetched = 0;
            PassengerEvent event;
            
            eventsFile.seek(metadata.syncedEvents * sizeof(PassengerEvent));
            
            while (eventsFile.available() >= sizeof(PassengerEvent) && fetched < maxCount) {
                size_t bytesRead = eventsFile.readBytes((char*)&event, sizeof(PassengerEvent));
                if (bytesRead == sizeof(PassengerEvent)) {
                    buffer[fetched++] = event;
                } else {
                    break;
                }
            }
            
            eventsFile.close();
            return fetched;
        } else {
            int fetched = 0;
            for (int i = 0; i < memoryCount && fetched < maxCount; i++) {
                if (!memoryBuffer[i].isSynced) {
                    buffer[fetched++] = memoryBuffer[i];
                }
            }
            return fetched;
        }
    }

    // Позначення подій як синхронізованих
    void markSynced(uint32_t upToLocalId) {
        if (littleFSAvailable) {
            uint32_t newSyncedCount = 0;
            
            if (LittleFS.exists(eventsFilePath)) {
                File eventsFile = LittleFS.open(eventsFilePath, "r");
                if (eventsFile) {
                    PassengerEvent event;
                    while (eventsFile.available() >= sizeof(PassengerEvent)) {
                        size_t bytesRead = eventsFile.readBytes((char*)&event, sizeof(PassengerEvent));
                        if (bytesRead == sizeof(PassengerEvent) && event.localId <= upToLocalId) {
                            newSyncedCount++;
                        } else {
                            break;
                        }
                    }
                    eventsFile.close();
                }
            }
            
            metadata.syncedEvents = newSyncedCount;
            saveMetadata();
        } else {
            for (int i = 0; i < memoryCount; i++) {
                if (memoryBuffer[i].localId <= upToLocalId) {
                    memoryBuffer[i].isSynced = true;
                }
            }
        }
    }

    int getCount() const { 
        return littleFSAvailable ? metadata.totalEvents : memoryCount;
    }
    
    int getUnsyncedCount() const { 
        if (littleFSAvailable) {
            return metadata.totalEvents - metadata.syncedEvents;
        } else {
            int unsynced = 0;
            for (int i = 0; i < memoryCount; i++) {
                if (!memoryBuffer[i].isSynced) unsynced++;
            }
            return unsynced;
        }
    }

    void clear() {
        if (littleFSAvailable) {
            if (LittleFS.exists(eventsFilePath)) {
                LittleFS.remove(eventsFilePath);
            }
            
            metadata.nextLocalId = 1;
            metadata.totalEvents = 0;
            metadata.syncedEvents = 0;
            metadata.fileVersion++;
            
            saveMetadata();
        } else {
            memoryCount = 0;
            memoryNextId = 1;
        }
        
        Serial.println("[EventBuffer] Buffer cleared");
    }
    
    // Компактування буферу
    bool compactBuffer() {
        if (!littleFSAvailable) {
            // Для memory fallback - видаляємо синхронізовані
            int writeIdx = 0;
            for (int i = 0; i < memoryCount; i++) {
                if (!memoryBuffer[i].isSynced) {
                    if (writeIdx != i) {
                        memoryBuffer[writeIdx] = memoryBuffer[i];
                    }
                    writeIdx++;
                }
            }
            memoryCount = writeIdx;
            return true;
        }
        
        if (metadata.syncedEvents == 0) {
            return true;
        }
        
        String tempFilePath = "/events_temp.dat";
        
        File oldFile = LittleFS.open(eventsFilePath, "r");
        File newFile = LittleFS.open(tempFilePath, "w");
        
        if (!oldFile || !newFile) {
            if (oldFile) oldFile.close();
            if (newFile) newFile.close();
            return false;
        }
        
        oldFile.seek(metadata.syncedEvents * sizeof(PassengerEvent));
        
        uint8_t buffer[FILE_COPY_BUFFER_SIZE];
        while (oldFile.available()) {
            size_t bytesRead = oldFile.read(buffer, sizeof(buffer));
            newFile.write(buffer, bytesRead);
        }
        
        oldFile.close();
        newFile.close();
        
        LittleFS.remove(eventsFilePath);
        LittleFS.rename(tempFilePath, eventsFilePath);
        
        metadata.totalEvents -= metadata.syncedEvents;
        metadata.syncedEvents = 0;
        saveMetadata();
        
        return true;
    }
    
    // Статистика
    void printStorageStats() {
        if (littleFSAvailable) {
            size_t totalBytes = LittleFS.totalBytes();
            size_t usedBytes = LittleFS.usedBytes();
            
            Serial.printf("[EventBuffer] LittleFS: %u/%u bytes (%.1f%%)\n", 
                usedBytes, totalBytes, (float)usedBytes / totalBytes * 100.0);
            Serial.printf("[EventBuffer] Events: %u total, %u pending\n", 
                metadata.totalEvents, getUnsyncedCount());
        } else {
            Serial.printf("[EventBuffer] Memory: %d/%d events\n", memoryCount, MEMORY_BUFFER_SIZE);
            Serial.printf("[EventBuffer] Pending: %d\n", getUnsyncedCount());
        }
    }
};

#endif
