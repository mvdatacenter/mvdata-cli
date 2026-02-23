package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
)

func TestInstanceTypesList(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, jsonHandler(http.StatusOK, []sdk.InstanceType{
			{InstanceType: "c1", HourlyPrice: 0.50},
			{InstanceType: "c1_g1", HourlyPrice: 1.25},
		})),
		"instance-types", "list",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	if gotPath != "/prices" {
		t.Errorf("expected /prices, got %s", gotPath)
	}
	if !strings.Contains(out, "c1") {
		t.Errorf("expected output to contain 'c1', got: %s", out)
	}
	if !strings.Contains(out, "c1_g1") {
		t.Errorf("expected output to contain 'c1_g1', got: %s", out)
	}
}

func TestInstanceTypesList_JSON(t *testing.T) {
	out, err := executeWithServer(
		jsonHandler(http.StatusOK, []sdk.InstanceType{
			{InstanceType: "c1", HourlyPrice: 0.50},
			{InstanceType: "c1_g1", HourlyPrice: 1.25},
		}),
		"instance-types", "list", "-o", "json",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var types []sdk.InstanceType
	if err := json.Unmarshal([]byte(out), &types); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(types) != 2 {
		t.Fatalf("expected 2 types, got %d", len(types))
	}
	if types[0].InstanceType != "c1" {
		t.Errorf("expected first type c1, got %s", types[0].InstanceType)
	}
}
