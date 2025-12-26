package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Простий тест для перевірки API системних налаштувань
func main() {
	baseURL := "http://localhost:8080/api"

	// Тестові дані для входу
	loginData := map[string]string{
		"email":    "admin@busoptima.ua",
		"password": "admin123",
	}

	// Логін для отримання токена
	token, err := login(baseURL, loginData)
	if err != nil {
		log.Fatal("Login failed:", err)
	}

	fmt.Println("✅ Login successful, token received")

	// Тест отримання поточних налаштувань
	fmt.Println("\n📋 Testing GET /admin/settings...")
	if err := testGetSettings(baseURL, token); err != nil {
		log.Printf("❌ GET settings failed: %v", err)
	} else {
		fmt.Println("✅ GET settings successful")
	}

	// Тест оновлення налаштувань
	fmt.Println("\n🔧 Testing PUT /admin/settings...")
	if err := testUpdateSettings(baseURL, token); err != nil {
		log.Printf("❌ PUT settings failed: %v", err)
	} else {
		fmt.Println("✅ PUT settings successful")
	}

	// Тест динамічного ціноутворення з новими налаштуваннями
	fmt.Println("\n💰 Testing POST /pricing/calculate...")
	if err := testPricing(baseURL, token); err != nil {
		log.Printf("❌ Pricing test failed: %v", err)
	} else {
		fmt.Println("✅ Pricing test successful")
	}
}

func login(baseURL string, loginData map[string]string) (string, error) {
	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if token, ok := result["access_token"].(string); ok {
		return token, nil
	}

	return "", fmt.Errorf("no token in response")
}

func testGetSettings(baseURL, token string) error {
	req, _ := http.NewRequest("GET", baseURL+"/admin/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var settings map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&settings)

	fmt.Printf("Current settings: %+v\n", settings)
	return nil
}

func testUpdateSettings(baseURL, token string) error {
	newSettings := map[string]interface{}{
		"fuel_price_per_liter":   52.50,
		"peak_hours_coefficient": 1.30,
		"weekend_coefficient":    1.20,
		"high_demand_threshold":  85,
		"low_demand_threshold":   30,
		"price_min_coefficient":  0.65,
		"price_max_coefficient":  1.60,
		"seasonal_coefficients": map[string]float64{
			"new_year": 1.35,
			"summer":   1.20,
			"regular":  1.00,
		},
	}

	jsonData, _ := json.Marshal(newSettings)
	req, _ := http.NewRequest("PUT", baseURL+"/admin/settings", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("Update result: %+v\n", result)
	return nil
}

func testPricing(baseURL, token string) error {
	pricingData := map[string]interface{}{
		"base_price":         200.00,
		"current_passengers": 35,
		"capacity":           50,
		"departure_time":     time.Now().Add(2 * time.Hour).Format(time.RFC3339),
	}

	jsonData, _ := json.Marshal(pricingData)
	req, _ := http.NewRequest("POST", baseURL+"/pricing/calculate", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("Pricing result: %+v\n", result)
	return nil
}
