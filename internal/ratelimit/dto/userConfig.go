package dto

type UserConfig struct {
	Ip         string  `json:"ip" bd:"ip"`
	Capacity   int     `json:"capacity" bd:"capacity"`
	RatePerSec float64 `json:"rate_per_sec" bd:"rate_per_sec"`
}
