package models

import (
	"testing"
	"time"
)

func TestEnvStruct(t *testing.T) {
	e := Env{
		ID:        1,
		Name:      "Node.js",
		Key:       "nodejs",
		Category:  CategoryRuntime,
		Icon:      "⚡",
		IsManaged: true,
		CreatedAt: time.Now(),
	}
	if e.Key != "nodejs" {
		t.Errorf("expected key 'nodejs', got %s", e.Key)
	}
	if e.Category != CategoryRuntime {
		t.Errorf("expected category runtime, got %s", e.Category)
	}
}

func TestHealthLevel(t *testing.T) {
	levels := []HealthLevel{HealthHealthy, HealthInfo, HealthWarning, HealthCritical}
	for _, l := range levels {
		if string(l) == "" {
			t.Error("health level should not be empty")
		}
	}
}

func TestEnvSummary(t *testing.T) {
	s := EnvSummary{
		Env: Env{
			Name: "Node.js",
			Key:  "nodejs",
		},
		TotalSize: 2 * 1024 * 1024 * 1024,
		Health:    HealthWarning,
	}
	if s.TotalSize != 2147483648 {
		t.Errorf("expected 2GB, got %d", s.TotalSize)
	}
	if s.Health != HealthWarning {
		t.Errorf("expected warning health")
	}
}
