package gofish

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestServer builds an httptest.Server that serves the given path ->
// JSON body map, and returns a Client bound to it.
func newTestServer(t *testing.T, routes map[string]interface{}) (*Client, *httptest.Server) {
	t.Helper()
	mux := http.NewServeMux()
	for path, body := range routes {
		body := body
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(body); err != nil {
				t.Fatalf("encode response: %v", err)
			}
		})
	}
	srv := httptest.NewServer(mux)

	c, err := NewClient(ClientConfig{
		Endpoint: srv.URL,
		Username: "admin",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, srv
}

func TestNewClientRequiresEndpoint(t *testing.T) {
	if _, err := NewClient(ClientConfig{}); err == nil {
		t.Fatal("expected error for empty endpoint, got nil")
	}
}

func TestGetServiceRoot(t *testing.T) {
	c, srv := newTestServer(t, map[string]interface{}{
		"/redfish/v1/": map[string]interface{}{
			"@odata.id":      "/redfish/v1/",
			"Id":             "RootService",
			"Name":           "Root Service",
			"RedfishVersion": "1.6.0",
			"Systems":        map[string]string{"@odata.id": "/redfish/v1/Systems"},
			"Chassis":        map[string]string{"@odata.id": "/redfish/v1/Chassis"},
			"Managers":       map[string]string{"@odata.id": "/redfish/v1/Managers"},
		},
	})
	defer srv.Close()

	sr, err := GetServiceRoot(context.Background(), c)
	if err != nil {
		t.Fatalf("GetServiceRoot: %v", err)
	}
	if sr.RedfishVersion != "1.6.0" {
		t.Errorf("RedfishVersion = %q, want %q", sr.RedfishVersion, "1.6.0")
	}
	if sr.Systems.ODataID != "/redfish/v1/Systems" {
		t.Errorf("Systems.ODataID = %q, want %q", sr.Systems.ODataID, "/redfish/v1/Systems")
	}
}

func TestListAndGetSystem(t *testing.T) {
	c, srv := newTestServer(t, map[string]interface{}{
		"/redfish/v1/Systems": map[string]interface{}{
			"Members@odata.count": 1,
			"Members": []map[string]string{
				{"@odata.id": "/redfish/v1/Systems/1"},
			},
		},
		"/redfish/v1/Systems/1": map[string]interface{}{
			"@odata.id":    "/redfish/v1/Systems/1",
			"Id":           "1",
			"Name":         "Computer System",
			"Manufacturer": "HPE",
			"PowerState":   "On",
			"Status":       map[string]string{"State": "Enabled", "Health": "OK"},
		},
	})
	defer srv.Close()

	ids, err := ListSystems(context.Background(), c)
	if err != nil {
		t.Fatalf("ListSystems: %v", err)
	}
	if len(ids) != 1 || ids[0] != "/redfish/v1/Systems/1" {
		t.Fatalf("ListSystems = %v, want 1 member /redfish/v1/Systems/1", ids)
	}

	sys, err := GetSystem(context.Background(), c, ids[0])
	if err != nil {
		t.Fatalf("GetSystem: %v", err)
	}
	if sys.Manufacturer != "HPE" {
		t.Errorf("Manufacturer = %q, want HPE", sys.Manufacturer)
	}
	if sys.PowerState != "On" {
		t.Errorf("PowerState = %q, want On", sys.PowerState)
	}
	if sys.Status.Health != "OK" {
		t.Errorf("Status.Health = %q, want OK", sys.Status.Health)
	}

	sysByID, err := GetSystem(context.Background(), c, "1")
	if err != nil {
		t.Fatalf("GetSystem by bare id: %v", err)
	}
	if sysByID.ID != "1" {
		t.Errorf("Id = %q, want 1", sysByID.ID)
	}
}

func TestSystemReset(t *testing.T) {
	var gotBody map[string]string
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Systems/1/Actions/ComputerSystem.Reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := NewClient(ClientConfig{Endpoint: srv.URL})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	sys := &ComputerSystem{ODataID: "/redfish/v1/Systems/1"}
	if err := sys.Reset(context.Background(), c, ResetTypeGracefulRestart); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if gotBody["ResetType"] != string(ResetTypeGracefulRestart) {
		t.Errorf("ResetType = %q, want %q", gotBody["ResetType"], ResetTypeGracefulRestart)
	}
}

func TestListAndGetChassis(t *testing.T) {
	c, srv := newTestServer(t, map[string]interface{}{
		"/redfish/v1/Chassis": map[string]interface{}{
			"Members@odata.count": 1,
			"Members": []map[string]string{
				{"@odata.id": "/redfish/v1/Chassis/1"},
			},
		},
		"/redfish/v1/Chassis/1": map[string]interface{}{
			"@odata.id":    "/redfish/v1/Chassis/1",
			"Id":           "1",
			"Name":         "Computer System Chassis",
			"ChassisType":  "RackMount",
			"Manufacturer": "HPE",
			"Power":        map[string]string{"@odata.id": "/redfish/v1/Chassis/1/Power"},
			"Thermal":      map[string]string{"@odata.id": "/redfish/v1/Chassis/1/Thermal"},
		},
		"/redfish/v1/Chassis/1/Power": map[string]interface{}{
			"@odata.id": "/redfish/v1/Chassis/1/Power",
			"Id":        "Power",
			"Name":      "Power",
			"PowerSupplies": []map[string]interface{}{
				{"Name": "PSU1", "PowerCapacityWatts": 800.0, "Status": map[string]string{"Health": "OK"}},
			},
		},
		"/redfish/v1/Chassis/1/Thermal": map[string]interface{}{
			"@odata.id": "/redfish/v1/Chassis/1/Thermal",
			"Id":        "Thermal",
			"Name":      "Thermal",
			"Fans": []map[string]interface{}{
				{"Name": "Fan 1", "Reading": 55, "ReadingUnits": "Percent"},
			},
			"Temperatures": []map[string]interface{}{
				{"Name": "CPU1", "ReadingCelsius": 42.0},
			},
		},
	})
	defer srv.Close()

	ids, err := ListChassis(context.Background(), c)
	if err != nil {
		t.Fatalf("ListChassis: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("ListChassis = %v, want 1 member", ids)
	}

	ch, err := GetChassis(context.Background(), c, ids[0])
	if err != nil {
		t.Fatalf("GetChassis: %v", err)
	}
	if ch.ChassisType != "RackMount" {
		t.Errorf("ChassisType = %q, want RackMount", ch.ChassisType)
	}

	pw, err := ch.GetPower(context.Background(), c)
	if err != nil {
		t.Fatalf("GetPower: %v", err)
	}
	if len(pw.PowerSupplies) != 1 || pw.PowerSupplies[0].Name != "PSU1" {
		t.Fatalf("PowerSupplies = %+v, want 1 entry named PSU1", pw.PowerSupplies)
	}

	th, err := ch.GetThermal(context.Background(), c)
	if err != nil {
		t.Fatalf("GetThermal: %v", err)
	}
	if len(th.Fans) != 1 || th.Fans[0].Name != "Fan 1" {
		t.Fatalf("Fans = %+v, want 1 entry named Fan 1", th.Fans)
	}
	if len(th.Temperatures) != 1 || th.Temperatures[0].ReadingCelsius != 42.0 {
		t.Fatalf("Temperatures = %+v, want 1 entry at 42.0C", th.Temperatures)
	}
}

func TestGetChassisNoPower(t *testing.T) {
	ch := &Chassis{ID: "1"}
	c, err := NewClient(ClientConfig{Endpoint: "https://example.invalid"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := ch.GetPower(context.Background(), c); err == nil {
		t.Fatal("expected error for chassis with no Power resource")
	}
}

func TestHTTPErrorOnNon2xx(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Systems/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := NewClient(ClientConfig{Endpoint: srv.URL})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = GetSystem(context.Background(), c, "missing")
	if err == nil {
		t.Fatal("expected HTTP error, got nil")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", httpErr.StatusCode)
	}
}
