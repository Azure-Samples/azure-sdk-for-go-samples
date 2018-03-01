// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package helpers

import (
	"encoding/json"
	"net/http"
	"time"
)

// audience represents list of audience endpoints
type audience []string

// authentication represents authentication section of Environment Information
type authentication struct {
	LoginEndpoint string   `json:"loginEndpoint"`
	Audiences     audience `json:"audiences"`
}

// EnvironmentInformation represents a set of endpoints for the of Azure's Environment.
type environmentInformation struct {
	GalleryEndpoint string         `json:"galleryEndpoint"`
	GraphEndpoint   string         `json:"graphEndpoint"`
	PortalEndpoint  string         `json:"portalEndpoint"`
	Authentication  authentication `json:"authentication"`
}

var client = &http.Client{Timeout: 3 * time.Second}

// GetAadResourceID retrieves AadResourceId from ARMEndpoint
func GetAadResourceID(armEndpointString string) (aadResourceID, aadEndpoint string, err error) {
	managementEndpoint := armEndpointString + "/metadata/endpoints?api-version=1.0"
	env := new(environmentInformation)
	if err := getJSON(managementEndpoint, env); err != nil {
		return aadResourceID, aadEndpoint, err
	}
	aadResourceID = env.Authentication.Audiences[0]
	aadEndpoint = env.Authentication.LoginEndpoint
	return aadResourceID, aadEndpoint, nil
}

// getJSON retrieves EnvironmentInformation
func getJSON(url string, target interface{}) error {
	response, err := client.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return json.NewDecoder(response.Body).Decode(target)
}
