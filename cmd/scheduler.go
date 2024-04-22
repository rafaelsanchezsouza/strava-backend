package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Activity struct {
	ResourceState      int     `json:"resource_state"`
	Athlete            Athlete `json:"athlete"`
	Name               string  `json:"name"`
	Distance           float64 `json:"distance"`
	MovingTime         int     `json:"moving_time"`
	ElapsedTime        int     `json:"elapsed_time"`
	TotalElevationGain float64 `json:"total_elevation_gain"`
	Type               string  `json:"type"`
	SportType          string  `json:"sport_type"`
	WorkoutType        *int    `json:"workout_type"` // Pointer to handle null
}

type Athlete struct {
	ResourceState int    `json:"resource_state"`
	Firstname     string `json:"firstname"`
	Lastname      string `json:"lastname"`
}

type RefreshTokenResponse struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	ExpiresAt    int    `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}

func refreshToken() (RefreshTokenResponse, error) {
	baseURL := "https://www.strava.com/oauth/token"
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	refreshToken := os.Getenv("REFRESH_TOKEN")
	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("client_secret", clientSecret)
	params.Add("grant_type", "refresh_token")
	params.Add("refresh_token", refreshToken)

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return RefreshTokenResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request: ", err)
		return RefreshTokenResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server returned non-200 status: %d %s", resp.StatusCode, resp.Status)
		return RefreshTokenResponse{}, err
	}

	var userResp RefreshTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		fmt.Println("Error decoding response: ", err)
		return RefreshTokenResponse{}, err
	}

	return userResp, nil
}

func fetchActivities(refreshTokenResponse RefreshTokenResponse) {

	clubID := os.Getenv("CLUB_ID")
	reqURL := fmt.Sprintf("https://www.strava.com/api/v3/clubs/%s/activities", clubID)

	token := refreshTokenResponse.TokenType + " " + refreshTokenResponse.AccessToken

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request: ", err)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var activities []Activity
	if err := json.Unmarshal(body, &activities); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	for _, activity := range activities {
		fmt.Printf("Activity: %s, Type: %s, Athlete: %s %s\n", activity.Name, activity.Type, activity.Athlete.Firstname, activity.Athlete.Lastname)
	}

}

// func main() {
// 	fmt.Println("Scheduler started, fetching activities once per day...")

// 	ticker := time.NewTicker(24 * time.Hour)
// 	defer ticker.Stop()

// 	fetchActivities() // Initial fetch

// 	for {
// 		select {
// 		case <-ticker.C:
// 			fetchActivities()
// 		}
// 	}
// }

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
	}
}

func main() {
	loadEnv()
	fmt.Println("Test started, fetching activities from club...")

	token, err := refreshToken()

	if err != nil {
		// Handle the error appropriately
		fmt.Println("Error refreshing token:", err)
		os.Exit(1) // Exit or handle the error as needed
	}

	fetchActivities(token)
}
