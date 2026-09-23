package gofish

import "context"

// Status is the common Redfish Status object embedded in most resources.
type Status struct {
	State        string `json:"State,omitempty"`
	Health       string `json:"Health,omitempty"`
	HealthRollup string `json:"HealthRollup,omitempty"`
}

// ServiceRoot represents the Redfish service root document at
// /redfish/v1/.
type ServiceRoot struct {
	ODataID        string  `json:"@odata.id"`
	ID             string  `json:"Id"`
	Name           string  `json:"Name"`
	RedfishVersion string  `json:"RedfishVersion"`
	Product        string  `json:"Product,omitempty"`
	Vendor         string  `json:"Vendor,omitempty"`
	UUID           string  `json:"UUID,omitempty"`
	Systems        ODataID `json:"Systems"`
	Chassis        ODataID `json:"Chassis"`
	Managers       ODataID `json:"Managers"`
}

// GetServiceRoot retrieves the Redfish service root document.
func GetServiceRoot(ctx context.Context, c *Client) (*ServiceRoot, error) {
	var sr ServiceRoot
	if err := c.Get(ctx, "/redfish/v1/", &sr); err != nil {
		return nil, err
	}
	return &sr, nil
}

// collection is the generic Redfish collection envelope used to enumerate
// members of resource collections (Systems, Chassis, Managers, ...).
type collection struct {
	MembersCount int       `json:"Members@odata.count"`
	Members      []ODataID `json:"Members"`
}

// getCollectionIDs fetches the collection at path and returns the
// @odata.id of each member.
func getCollectionIDs(ctx context.Context, c *Client, path string) ([]string, error) {
	var col collection
	if err := c.Get(ctx, path, &col); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(col.Members))
	for _, m := range col.Members {
		ids = append(ids, m.ODataID)
	}
	return ids, nil
}
