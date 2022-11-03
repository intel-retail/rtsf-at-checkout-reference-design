// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package rfidgtin

import (
	"strconv"
)

const (
	EPCPureURIPrefix = "urn:epc:id:sgtin:"
)

type TagDecoder interface {
	Decode(tagData string) (productID, URI string, err error)
	Type() string
}

func GetEpcBytes(epc string) ([numEpcBytes]byte, error) {
	epcBytes := [numEpcBytes]byte{}
	for i := 0; i < len(epcBytes); i++ {
		tempParse, err := strconv.ParseUint(epc[i*2:(i*2)+2], 16, 8)
		if err != nil {
			return epcBytes, err
		}
		epcBytes[i] = byte(tempParse) & 0xFF
	}
	return epcBytes, nil
}

func ZeroFill(data string, num int) string {
	for {
		if len(data) >= num {
			return data[0:num]
		}
		data = "0" + data
	}
}
