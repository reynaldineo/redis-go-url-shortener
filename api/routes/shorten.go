package routes

import "time"

type (
	request struct {
		URL         string        `json:"url"`
		CustomShort string        `json:"custom_short"`
		Expiry      time.Duration `json:"expiry"`
	}

	response struct {
		URL            string        `json:"url"`
		CustomShort    string        `json:"custom_short"`
		Expiry         time.Duration `json:"expiry"`
		XRateRemaining int           `json:"x-rate-remaining"`
		XRateLimitRest time.Duration `json:"x-rate-limit-rest"`
	}
)
