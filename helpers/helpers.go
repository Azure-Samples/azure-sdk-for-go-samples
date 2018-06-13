// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/Azure/go-autorest/autorest/utils"
)

// PrintAndLog writes to stdout and to a logger.
func PrintAndLog(message string) {
	log.Println(message)
	fmt.Println(message)
}

func contains(array []string, element string) bool {
	for _, e := range array {
		if e == element {
			return true
		}
	}
	return false
}

// UserAgent return the string to be appended to user agent header
func UserAgent() string {
	return "samples " + utils.GetCommit()
}

// ReadJSON reads a json file, and unmashals it.
// Very useful for template deployments.
func ReadJSON(path string) (*map[string]interface{}, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read template file: %v\n", err)
	}
	contents := make(map[string]interface{})
	json.Unmarshal(data, &contents)
	return &contents, nil
}
