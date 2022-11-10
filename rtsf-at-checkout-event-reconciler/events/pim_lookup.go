// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

func productLookup(productID string, productLookupEndpoint string) (ProductDetails, error) {

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
