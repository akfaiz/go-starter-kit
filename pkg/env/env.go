// Package env provides small helpers to read environment variables
// with consistent defaults and parsing.
package env

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// ---------- Lookup variants (report presence + parse error) ----------

func LookupString(name string) (string, bool) {
	return os.LookupEnv(name)
}

func LookupInt(name string) (int, bool, error) {
	v, ok := os.LookupEnv(name)
	if !ok {
		return 0, false, nil
	}
	i, err := strconv.Atoi(strings.TrimSpace(v))
	return i, true, err
}

func LookupBool(name string) (bool, bool, error) {
	v, ok := os.LookupEnv(name)
	if !ok {
		return false, false, nil
	}
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	return b, true, err
}

func LookupFloat(name string) (float64, bool, error) {
	v, ok := os.LookupEnv(name)
	if !ok {
		return 0, false, nil
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	return f, true, err
}

func LookupDuration(name string) (time.Duration, bool, error) {
	v, ok := os.LookupEnv(name)
	if !ok {
		return 0, false, nil
	}
	d, err := time.ParseDuration(strings.TrimSpace(v))
	return d, true, err
}

// ---------- High-level getters (with optional default) ----------

func GetString(name string, defaultValue ...string) string {
	if v, ok := LookupString(name); ok {
		return v
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

func GetInt(name string, defaultValue ...int) int {
	if v, ok, _ := LookupInt(name); ok {
		return v
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

func GetBool(name string, defaultValue ...bool) bool {
	if v, ok, _ := LookupBool(name); ok {
		return v
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

func GetFloat(name string, defaultValue ...float64) float64 {
	if v, ok, _ := LookupFloat(name); ok {
		return v
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

func GetDuration(name string, defaultValue ...time.Duration) time.Duration {
	if v, ok, _ := LookupDuration(name); ok {
		return v
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// ---------- Must* variants (panic on missing/invalid) ----------

func MustGetString(name string) string {
	v, ok := os.LookupEnv(name)
	if !ok {
		panic("env: missing " + name)
	}
	return v
}

func MustGetInt(name string) int {
	i, ok, err := LookupInt(name)
	if !ok {
		panic("env: missing " + name)
	}
	if err != nil {
		panic("env: invalid int for " + name + ": " + err.Error())
	}
	return i
}

func MustGetBool(name string) bool {
	b, ok, err := LookupBool(name)
	if !ok {
		panic("env: missing " + name)
	}
	if err != nil {
		panic("env: invalid bool for " + name + ": " + err.Error())
	}
	return b
}

func MustGetFloat(name string) float64 {
	f, ok, err := LookupFloat(name)
	if !ok {
		panic("env: missing " + name)
	}
	if err != nil {
		panic("env: invalid float for " + name + ": " + err.Error())
	}
	return f
}

func MustGetDuration(name string) time.Duration {
	d, ok, err := LookupDuration(name)
	if !ok {
		panic("env: missing " + name)
	}
	if err != nil {
		panic("env: invalid duration for " + name + ": " + err.Error())
	}
	return d
}
