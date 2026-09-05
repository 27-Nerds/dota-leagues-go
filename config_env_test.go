package main

import "testing"

func TestConfigEnvironmentOverrides(t *testing.T) {
	t.Setenv("DOTA_DEVELOPMENT_DATABASE_URL", "http://database:8529")
	t.Setenv("DOTA_DEVELOPMENT_VALVE_RPS", "0.75")
	t.Setenv("DOTA_DEVELOPMENT_THROTTLING_PER_MIN", "120")
	if GetConfigStr("database.url") != "http://database:8529" || GetConfigFloat("valve.rps") != 0.75 || GetConfigInt("throttling.per_min") != 120 {
		t.Fatal("environment configuration not applied")
	}
}
