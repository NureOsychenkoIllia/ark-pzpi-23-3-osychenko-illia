#ifndef PRICING_ENGINE_H
#define PRICING_ENGINE_H

#include "config.h"
#include "models.h"
#include <Arduino.h>
#include <time.h>

class PricingEngine {
private:
    // Розрахунок коефіцієнта попиту на основі завантаженості
    float calculateDemandCoefficient(float occupancyRate) {
        if (occupancyRate < OCCUPANCY_LOW_THRESHOLD) return DEMAND_COEFF_LOW;
        if (occupancyRate < OCCUPANCY_MEDIUM_THRESHOLD) return DEMAND_COEFF_MEDIUM;
        if (occupancyRate < OCCUPANCY_HIGH_THRESHOLD) return DEMAND_COEFF_HIGH;
        return DEMAND_COEFF_VERY_HIGH;
    }

    // Розрахунок коефіцієнта часу
    float calculateTimeCoefficient(int hour) {
        // Пікові години
        if ((hour >= PEAK_MORNING_START && hour <= PEAK_MORNING_END) || 
            (hour >= PEAK_EVENING_START && hour <= PEAK_EVENING_END)) {
            return TIME_COEFF_PEAK;
        }
        // Нічні години
        if (hour >= NIGHT_START || hour <= NIGHT_END) {
            return TIME_COEFF_NIGHT;
        }
        return TIME_COEFF_NORMAL;
    }

    // Розрахунок коефіцієнта дня тижня
    float calculateDayCoefficient(int dayOfWeek) {
        // 0 = неділя, 6 = субота
        if (dayOfWeek == 0 || dayOfWeek == 6) {
            return DAY_COEFF_WEEKEND;
        }
        return DAY_COEFF_WEEKDAY;
    }

    // Округлення ціни до найближчого кроку
    float roundPrice(float price) {
        return round(price / PRICE_ROUND_STEP) * PRICE_ROUND_STEP;
    }

public:
    // Розрахунок рекомендованої ціни
    PriceRecommendation calculatePrice(float basePrice, int currentPassengers, 
                                        int capacity, struct tm* timeinfo) {
        PriceRecommendation rec;
        rec.basePrice = basePrice;
        rec.calculatedAt = millis();

        // Розрахунок завантаженості
        rec.occupancyRate = (capacity > 0) 
            ? (float)currentPassengers / capacity * 100.0 
            : 0.0;

        // Розрахунок коефіцієнтів
        rec.demandCoeff = calculateDemandCoefficient(rec.occupancyRate);
        rec.timeCoeff = calculateTimeCoefficient(timeinfo ? timeinfo->tm_hour : 12);
        rec.dayCoeff = calculateDayCoefficient(timeinfo ? timeinfo->tm_wday : 1);

        // Розрахунок рекомендованої ціни
        float rawPrice = basePrice * rec.demandCoeff * rec.timeCoeff * rec.dayCoeff;

        // Обмеження діапазону
        float minPrice = basePrice * PRICE_MIN_COEFF;
        float maxPrice = basePrice * PRICE_MAX_COEFF;
        rawPrice = constrain(rawPrice, minPrice, maxPrice);

        // Округлення
        rec.recommendedPrice = roundPrice(rawPrice);

        Serial.printf("[Pricing] Base=%.2f, Load=%.1f%%, K_demand=%.2f, K_time=%.2f, K_day=%.2f -> Rec=%.2f\n",
            basePrice, rec.occupancyRate, rec.demandCoeff, rec.timeCoeff, rec.dayCoeff, rec.recommendedPrice);

        return rec;
    }

    // Отримання текстової категорії ціни
    const char* getPriceCategory(float basePrice, float recommendedPrice) {
        float ratio = recommendedPrice / basePrice;
        if (ratio < PRICE_CATEGORY_DISCOUNT) return "DISCOUNT";
        if (ratio < PRICE_CATEGORY_LOW) return "LOW";
        if (ratio < PRICE_CATEGORY_NORMAL) return "NORMAL";
        if (ratio < PRICE_CATEGORY_HIGH) return "HIGH";
        return "PEAK";
    }
};

#endif
