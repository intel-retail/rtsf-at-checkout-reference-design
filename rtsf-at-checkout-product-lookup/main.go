// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var products = make(map[string]ProductInfo)

type ProductInfo struct {
	Barcode      string  `json:"barcode"`
	Name         string  `json:"name"`
	MinWeight    float64 `json:"min_weight"`
	MaxWeight    float64 `json:"max_weight"`
	RfidEligible bool    `json:"rfid_eligible"`
}

func main() {
	localDatabase := ""
	flag.StringVar(&localDatabase, "file", "", "name of file with min/max weights")
	flag.Parse()

	if localDatabase != "" {
		setUpLocalDatabase(localDatabase)
		log.Print("Using JSON Database")
	} else {
		log.Print("No Database Specified")
		os.Exit(1)
	}

	initializeServer()
}

func weightLookupHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	params := mux.Vars(r)
	prodID := params["product_id"]

	log.Printf("Looking up product for productID: %v\n", prodID)

	var productInfo ProductInfo
	var err error

	if len(products) == 0 {

		log.Print("Error: Database has no information")
	} else {

		productInfo, err = localWeightLookupbyProductID(prodID)
	}

	if err != nil {
		log.Printf("%v: %v", err.Error(), prodID)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(productInfo)
}

func initializeServer() {

	port := os.Getenv("APP_PORT")

	if port == "" {
		port = "8083"
	}

	router := mux.NewRouter()

	router.HandleFunc("/weight/{product_id}", weightLookupHandler).Methods("GET")

	log.Printf("Product Lookup started listening on port: %s\n", port)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}

func localWeightLookupbyProductID(productID string) (ProductInfo, error) {
	productInfo, ok := products[productID]
	if !ok {
		return ProductInfo{}, errors.New("Could not find product in local database")
	}

	if productInfo.MinWeight == 0 && productInfo.MaxWeight == 0 {
		errString := "Product ID expected weight range not initialized"
		fmt.Println(errString)
		return productInfo, errors.New(errString)
	}

	return productInfo, nil
}

func setUpLocalDatabase(localDatabase string) {
	jsonFile, err := os.Open(localDatabase)
	defer jsonFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}

	jsonBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var productInfos []ProductInfo
	err = json.Unmarshal(jsonBytes, &productInfos)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//convert slice of productInfos to map[productID] = WeightInfo
	for _, product := range productInfos {
		products[product.Barcode] = ProductInfo{Name: product.Name, MinWeight: product.MinWeight, MaxWeight: product.MaxWeight, RfidEligible: product.RfidEligible}
	}
}
