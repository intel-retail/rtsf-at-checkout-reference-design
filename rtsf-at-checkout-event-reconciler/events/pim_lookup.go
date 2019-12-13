// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
)

func productLookup(productID string, edgexcontext *appcontext.Context) (ProductDetails, error) {

	appSettings := edgexcontext.Configuration.ApplicationSettings
	if appSettings == nil {
		edgexcontext.LoggingClient.Error("No application settings found")
		os.Exit(-1)
	}

	productLookupEndpoint, ok := appSettings["ProductLookupEndpoint"]
	if !ok {
		edgexcontext.LoggingClient.Error("ProductLookupEndpoint application setting not found")
		os.Exit(-1)
	}

	resp, err := http.Get("http://" + productLookupEndpoint + "/weight/" + productID)
	if err != nil {
		return ProductDetails{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errString, _ := ioutil.ReadAll(resp.Body)
		return ProductDetails{}, errors.New(string(errString))
	}

	var prodDetails ProductDetails
	err = json.NewDecoder(resp.Body).Decode(&prodDetails)
	if err != nil {
		return ProductDetails{}, err
	}

	return prodDetails, nil
}
