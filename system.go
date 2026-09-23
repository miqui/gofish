package gofish

import (
	"context"
	"fmt"
)

// ComputerSystem represents a Redfish ComputerSystem resource, the core
// "compute" resource for a physical or logical server (Systems/{id}).
type ComputerSystem struct {
	ODataID          string `json:"@odata.id"`
	ID               string `json:"Id"`
	Name             string `json:"Name"`
	SystemType       string `json:"SystemType,omitempty"`
	Manufacturer     string `json:"Manufacturer,omitempty"`
	Model            string `json:"Model,omitempty"`
	SKU              string `json:"SKU,omitempty"`
	SerialNumber     string `json:"SerialNumber,omitempty"`
	PartNumber       string `json:"PartNumber,omitempty"`
	UUID             string `json:"UUID,omitempty"`
	BiosVersion      string `json:"BiosVersion,omitempty"`
	PowerState       string `json:"PowerState,omitempty"`
	IndicatorLED     string `json:"IndicatorLED,omitempty"`
	Status           Status `json:"Status,omitempty"`
	ProcessorSummary struct {
		Count  int    `json:"Count,omitempty"`
		Model  string `json:"Model,omitempty"`
		Status Status `json:"Status,omitempty"`
	} `json:"ProcessorSummary,omitempty"`
	MemorySummary struct {
		TotalSystemMemoryGiB float64 `json:"TotalSystemMemoryGiB,omitempty"`
		Status               Status  `json:"Status,omitempty"`
	} `json:"MemorySummary,omitempty"`
	Processors         ODataID   `json:"Processors"`
	Memory             ODataID   `json:"Memory"`
	EthernetInterfaces ODataID   `json:"EthernetInterfaces"`
	Bios               ODataID   `json:"Bios"`
	Storage            ODataID   `json:"Storage"`
	Chassis            []ODataID `json:"Links,omitempty"`
}

// ResetType enumerates the values accepted by the ComputerSystem.Reset action.
type ResetType string

const (
	ResetTypeOn               ResetType = "On"
	ResetTypeForceOff         ResetType = "ForceOff"
	ResetTypeGracefulShutdown ResetType = "GracefulShutdown"
	ResetTypeGracefulRestart  ResetType = "GracefulRestart"
	ResetTypeForceRestart     ResetType = "ForceRestart"
	ResetTypeNmi              ResetType = "Nmi"
	ResetTypePushPowerButton  ResetType = "PushPowerButton"
	ResetTypeForceOn          ResetType = "ForceOn"
)

// ListSystems returns the @odata.id of every ComputerSystem exposed by the
// service's Systems collection.
func ListSystems(ctx context.Context, c *Client) ([]string, error) {
	return getCollectionIDs(ctx, c, "/redfish/v1/Systems")
}

// GetSystem retrieves a single ComputerSystem by its resource path (as
// returned by ListSystems) or by bare ID (e.g. "1").
func GetSystem(ctx context.Context, c *Client, idOrPath string) (*ComputerSystem, error) {
	path := idOrPath
	if !isODataPath(path) {
		path = fmt.Sprintf("/redfish/v1/Systems/%s", idOrPath)
	}
	var sys ComputerSystem
	if err := c.Get(ctx, path, &sys); err != nil {
		return nil, err
	}
	return &sys, nil
}

// Reset invokes the ComputerSystem.Reset action for this system.
func (s *ComputerSystem) Reset(ctx context.Context, c *Client, resetType ResetType) error {
	body := struct {
		ResetType ResetType `json:"ResetType"`
	}{ResetType: resetType}
	actionPath := s.ODataID + "/Actions/ComputerSystem.Reset"
	return c.Post(ctx, actionPath, body, nil)
}

// Processor represents a Redfish Processor resource
// (Systems/{id}/Processors/{id}).
type Processor struct {
	ODataID               string `json:"@odata.id"`
	ID                    string `json:"Id"`
	Name                  string `json:"Name"`
	Socket                string `json:"Socket,omitempty"`
	ProcessorType         string `json:"ProcessorType,omitempty"`
	ProcessorArchitecture string `json:"ProcessorArchitecture,omitempty"`
	Manufacturer          string `json:"Manufacturer,omitempty"`
	Model                 string `json:"Model,omitempty"`
	MaxSpeedMHz           int    `json:"MaxSpeedMHz,omitempty"`
	TotalCores            int    `json:"TotalCores,omitempty"`
	TotalThreads          int    `json:"TotalThreads,omitempty"`
	Status                Status `json:"Status,omitempty"`
}

// ListProcessors returns the @odata.id of every Processor under the given
// ComputerSystem.
func (s *ComputerSystem) ListProcessors(ctx context.Context, c *Client) ([]string, error) {
	return getCollectionIDs(ctx, c, s.Processors.ODataID)
}

// GetProcessor retrieves a single Processor by resource path.
func GetProcessor(ctx context.Context, c *Client, path string) (*Processor, error) {
	var p Processor
	if err := c.Get(ctx, path, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Memory represents a Redfish Memory resource (Systems/{id}/Memory/{id}),
// i.e. a single DIMM.
type Memory struct {
	ODataID           string `json:"@odata.id"`
	ID                string `json:"Id"`
	Name              string `json:"Name"`
	MemoryDeviceType  string `json:"MemoryDeviceType,omitempty"`
	Manufacturer      string `json:"Manufacturer,omitempty"`
	CapacityMiB       int    `json:"CapacityMiB,omitempty"`
	OperatingSpeedMhz int    `json:"OperatingSpeedMhz,omitempty"`
	SerialNumber      string `json:"SerialNumber,omitempty"`
	PartNumber        string `json:"PartNumber,omitempty"`
	Status            Status `json:"Status,omitempty"`
}

// ListMemory returns the @odata.id of every Memory (DIMM) resource under
// the given ComputerSystem.
func (s *ComputerSystem) ListMemory(ctx context.Context, c *Client) ([]string, error) {
	return getCollectionIDs(ctx, c, s.Memory.ODataID)
}

// GetMemory retrieves a single Memory resource by path.
func GetMemory(ctx context.Context, c *Client, path string) (*Memory, error) {
	var m Memory
	if err := c.Get(ctx, path, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// EthernetInterface represents a Redfish EthernetInterface resource
// (Systems/{id}/EthernetInterfaces/{id}).
type EthernetInterface struct {
	ODataID             string `json:"@odata.id"`
	ID                  string `json:"Id"`
	Name                string `json:"Name"`
	PermanentMACAddress string `json:"PermanentMACAddress,omitempty"`
	MACAddress          string `json:"MACAddress,omitempty"`
	SpeedMbps           int    `json:"SpeedMbps,omitempty"`
	InterfaceEnabled    bool   `json:"InterfaceEnabled,omitempty"`
	Status              Status `json:"Status,omitempty"`
}

// ListEthernetInterfaces returns the @odata.id of every EthernetInterface
// under the given ComputerSystem.
func (s *ComputerSystem) ListEthernetInterfaces(ctx context.Context, c *Client) ([]string, error) {
	return getCollectionIDs(ctx, c, s.EthernetInterfaces.ODataID)
}

// GetEthernetInterface retrieves a single EthernetInterface by path.
func GetEthernetInterface(ctx context.Context, c *Client, path string) (*EthernetInterface, error) {
	var e EthernetInterface
	if err := c.Get(ctx, path, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// Bios represents the Redfish Bios resource for a system
// (Systems/{id}/Bios), exposing vendor BIOS attributes as a free-form map.
type Bios struct {
	ODataID    string                 `json:"@odata.id"`
	ID         string                 `json:"Id"`
	Name       string                 `json:"Name"`
	Attributes map[string]interface{} `json:"Attributes,omitempty"`
}

// GetBios retrieves the Bios resource for the given ComputerSystem.
func (s *ComputerSystem) GetBios(ctx context.Context, c *Client) (*Bios, error) {
	var b Bios
	if err := c.Get(ctx, s.Bios.ODataID, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// isODataPath reports whether p looks like an absolute Redfish resource
// path (starting with "/redfish").
func isODataPath(p string) bool {
	return len(p) > 0 && p[0] == '/'
}
