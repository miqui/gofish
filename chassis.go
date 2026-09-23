package gofish

import (
	"context"
	"fmt"
)

// Chassis represents a Redfish Chassis resource, the physical enclosure
// housing compute, power, and thermal subsystems (Chassis/{id}).
type Chassis struct {
	ODataID         string  `json:"@odata.id"`
	ID              string  `json:"Id"`
	Name            string  `json:"Name"`
	ChassisType     string  `json:"ChassisType,omitempty"`
	Manufacturer    string  `json:"Manufacturer,omitempty"`
	Model           string  `json:"Model,omitempty"`
	SKU             string  `json:"SKU,omitempty"`
	SerialNumber    string  `json:"SerialNumber,omitempty"`
	PartNumber      string  `json:"PartNumber,omitempty"`
	AssetTag        string  `json:"AssetTag,omitempty"`
	IndicatorLED    string  `json:"IndicatorLED,omitempty"`
	Status          Status  `json:"Status,omitempty"`
	Power           ODataID `json:"Power"`
	Thermal         ODataID `json:"Thermal"`
	NetworkAdapters ODataID `json:"NetworkAdapters"`
}

// ListChassis returns the @odata.id of every Chassis exposed by the
// service's Chassis collection.
func ListChassis(ctx context.Context, c *Client) ([]string, error) {
	return getCollectionIDs(ctx, c, "/redfish/v1/Chassis")
}

// GetChassis retrieves a single Chassis by resource path or bare ID.
func GetChassis(ctx context.Context, c *Client, idOrPath string) (*Chassis, error) {
	path := idOrPath
	if !isODataPath(path) {
		path = fmt.Sprintf("/redfish/v1/Chassis/%s", idOrPath)
	}
	var ch Chassis
	if err := c.Get(ctx, path, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// PowerSupply represents a single entry in a Chassis Power resource's
// PowerSupplies array.
type PowerSupply struct {
	Name                 string  `json:"Name,omitempty"`
	MemberID             string  `json:"MemberId,omitempty"`
	PowerSupplyType      string  `json:"PowerSupplyType,omitempty"`
	LineInputVoltage     float64 `json:"LineInputVoltage,omitempty"`
	PowerCapacityWatts   float64 `json:"PowerCapacityWatts,omitempty"`
	LastPowerOutputWatts float64 `json:"LastPowerOutputWatts,omitempty"`
	Manufacturer         string  `json:"Manufacturer,omitempty"`
	Model                string  `json:"Model,omitempty"`
	SerialNumber         string  `json:"SerialNumber,omitempty"`
	Status               Status  `json:"Status,omitempty"`
}

// Power represents the Redfish Power resource for a chassis
// (Chassis/{id}/Power), aggregating power supply and control information.
type Power struct {
	ODataID       string        `json:"@odata.id"`
	ID            string        `json:"Id"`
	Name          string        `json:"Name"`
	PowerSupplies []PowerSupply `json:"PowerSupplies,omitempty"`
}

// GetPower retrieves the Power resource for the given Chassis.
func (ch *Chassis) GetPower(ctx context.Context, c *Client) (*Power, error) {
	if ch.Power.ODataID == "" {
		return nil, fmt.Errorf("gofish: chassis %s has no Power resource", ch.ID)
	}
	var p Power
	if err := c.Get(ctx, ch.Power.ODataID, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Fan represents a single entry in a Thermal resource's Fans array.
type Fan struct {
	Name         string `json:"Name,omitempty"`
	MemberID     string `json:"MemberId,omitempty"`
	Reading      int    `json:"Reading,omitempty"`
	ReadingUnits string `json:"ReadingUnits,omitempty"`
	Status       Status `json:"Status,omitempty"`
}

// Temperature represents a single entry in a Thermal resource's
// Temperatures array.
type Temperature struct {
	Name                   string  `json:"Name,omitempty"`
	MemberID               string  `json:"MemberId,omitempty"`
	ReadingCelsius         float64 `json:"ReadingCelsius,omitempty"`
	UpperThresholdCritical float64 `json:"UpperThresholdCritical,omitempty"`
	Status                 Status  `json:"Status,omitempty"`
}

// Thermal represents the Redfish Thermal resource for a chassis
// (Chassis/{id}/Thermal), aggregating fan and temperature sensor data.
type Thermal struct {
	ODataID      string        `json:"@odata.id"`
	ID           string        `json:"Id"`
	Name         string        `json:"Name"`
	Fans         []Fan         `json:"Fans,omitempty"`
	Temperatures []Temperature `json:"Temperatures,omitempty"`
}

// GetThermal retrieves the Thermal resource for the given Chassis.
func (ch *Chassis) GetThermal(ctx context.Context, c *Client) (*Thermal, error) {
	if ch.Thermal.ODataID == "" {
		return nil, fmt.Errorf("gofish: chassis %s has no Thermal resource", ch.ID)
	}
	var t Thermal
	if err := c.Get(ctx, ch.Thermal.ODataID, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Manager represents a Redfish Manager resource, i.e. the iLO management
// controller itself (Managers/{id}).
type Manager struct {
	ODataID         string `json:"@odata.id"`
	ID              string `json:"Id"`
	Name            string `json:"Name"`
	ManagerType     string `json:"ManagerType,omitempty"`
	Model           string `json:"Model,omitempty"`
	FirmwareVersion string `json:"FirmwareVersion,omitempty"`
	Status          Status `json:"Status,omitempty"`
}

// ListManagers returns the @odata.id of every Manager exposed by the
// service's Managers collection.
func ListManagers(ctx context.Context, c *Client) ([]string, error) {
	return getCollectionIDs(ctx, c, "/redfish/v1/Managers")
}

// GetManager retrieves a single Manager by resource path or bare ID.
func GetManager(ctx context.Context, c *Client, idOrPath string) (*Manager, error) {
	path := idOrPath
	if !isODataPath(path) {
		path = fmt.Sprintf("/redfish/v1/Managers/%s", idOrPath)
	}
	var m Manager
	if err := c.Get(ctx, path, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
