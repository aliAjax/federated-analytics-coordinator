package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Address         string
	ShutdownTimeout time.Duration
	MaxBodyBytes    int64
	MinParticipants int
	Environment     string
}

func Load() (Config, error) {
	c := Config{Address: get("FAC_ADDRESS", ":8081"), ShutdownTimeout: duration("FAC_SHUTDOWN_TIMEOUT", 15*time.Second), MaxBodyBytes: integer64("FAC_MAX_BODY_BYTES", 1<<20), MinParticipants: integer("FAC_MIN_PARTICIPANTS", 3), Environment: get("FAC_ENV", "development")}
	if c.MinParticipants < 2 {
		return Config{}, fmt.Errorf("minimum participants must be at least two")
	}
	if c.MaxBodyBytes < 1024 {
		return Config{}, fmt.Errorf("max body bytes must be at least 1024")
	}
	return c, nil
}
func get(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func integer(k string, d int) int {
	v, e := strconv.Atoi(get(k, strconv.Itoa(d)))
	if e != nil {
		return d
	}
	return v
}
func integer64(k string, d int64) int64 {
	v, e := strconv.ParseInt(get(k, strconv.FormatInt(d, 10)), 10, 64)
	if e != nil {
		return d
	}
	return v
}
func duration(k string, d time.Duration) time.Duration {
	v, e := time.ParseDuration(get(k, d.String()))
	if e != nil {
		return d
	}
	return v
}
