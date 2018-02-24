// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cosmosdb

import (
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"

	"github.com/Azure/go-autorest/autorest"
)

func getAuthorizer() (*autorest.BearerAuthorizer, error) {
	token, err := iam.GetResourceManagementToken(iam.AuthGrantType())

	if err != nil {
		return nil, fmt.Errorf("Failure to get management token: %s", err.Error())
	}

	return autorest.NewBearerAuthorizer(token), nil
}
