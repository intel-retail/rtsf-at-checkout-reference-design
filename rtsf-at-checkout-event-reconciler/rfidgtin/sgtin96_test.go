// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package rfidgtin

import (
	"strings"
	"testing"
)

//nolint :dupl
func testGtin14(t *testing.T, epc string, gtin string) {
	result, err := GetGtin14(epc)
	if err != nil {
		t.Errorf("[FAIL] unable to convert epc %s to gtin14: %s", epc, err.Error())
	} else if result != gtin {
		t.Errorf("[FAIL] expected gtin: %s, but got %s for epc %s", gtin, result, epc)
	} else {
		t.Logf(`[PASS] epc: %s, gtin14: %s`, epc, gtin)
	}
}

func testGtin14Error(t *testing.T, epc string, s string) {
	gtin, err := GetGtin14(epc)
	if err == nil {
		t.Errorf("[FAIL] expecting an error converting epc %s to gtin14, but got: %s", epc, gtin)
	} else if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(s)) {
		t.Errorf(`[FAIL] expected error to contain "%s", but got "%s"`, s, err.Error())
	} else {
		t.Logf(`[PASS] expected error string "%s" was successfully found in error "%s"`, s, err.Error())
	}
}

//nolint :dupl
func testCompanyPrefix(t *testing.T, epc string, expected int64) {
	result, err := GetCompanyPrefixByEpc(epc)
	if err != nil {
		t.Errorf("[FAIL] unable to convert epc %s to gtin14: %s", epc, err.Error())
	} else if result != expected {
		t.Errorf("[FAIL] expected company prefix: %d, but got %d for epc %s", expected, result, epc)
	} else {
		t.Logf(`[PASS] epc: %s, company prefix: %d`, epc, expected)
	}
}

func testItemFilter(t *testing.T, epc string, expected int64) {
	result, err := GetItemFilter(epc)
	if err != nil {
		t.Errorf("[FAIL] unable to get filter value")
	} else if result != expected {
		t.Errorf("[FAIL] filter value incorrect")
	} else {
		t.Log("[PASS]")
	}
}

//nolint:unparam
func testItemFilterError(t *testing.T, epc string, expectedErrorMessage string) {
	result, err := GetItemFilter(epc)
	if err == nil {
		t.Errorf("[FAIL] expecting an error but got %d", result)
	} else if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(expectedErrorMessage)) {
		t.Errorf(`[FAIL] expected error to contain "%s", but got "%s"`, expectedErrorMessage, err.Error())
	} else {
		t.Log("[PASS]")
	}
}

func TestItemFilter0(t *testing.T) {
	testItemFilter(t, "3000011B896A506B29C18539", 0)
}
func TestItemFilter1(t *testing.T) {
	testItemFilter(t, "302000662D3D311048C6D8D9", 1)
}
func TestItemFilter2(t *testing.T) {
	testItemFilter(t, "304000662D3D311048C6D8D9", 2)
}
func TestItemFilter3(t *testing.T) {
	testItemFilter(t, "306000662D3D311048C6D8D9", 3)
}
func TestItemFilter4(t *testing.T) {
	testItemFilter(t, "308000662D3D311048C6D8D9", 4)
}
func TestItemFilter5(t *testing.T) {
	testItemFilter(t, "30B100662D3D311048C6D8D9", 5)
}
func TestItemFilter6(t *testing.T) {
	testItemFilter(t, "30D100662D3D311048C6D8D9", 6)
}
func TestItemFilter7(t *testing.T) {
	testItemFilter(t, "30F100662D3D311048C6D8D9", 7)
}
func TestInvalidItemFilter7(t *testing.T) {
	testItemFilterError(t, "30A02AC002E4789", "Not a properly encoded SGTIN")
}
func TestGetGtin14Partition0(t *testing.T) {
	// Partition Value = 0
	testGtin14(t, "300000000000044000000001", "10000000000014")
	testGtin14(t, "300000662D3D311048C6D8D9", "40004285602049")
	testGtin14(t, "3000011B896A506B29C18539", "10011892394440")
}

func TestGetGtin14Partition1(t *testing.T) {
	// Partition Value = 1
	testGtin14(t, "300400000000204000000001", "00000000000116")
}

func TestGetGtin14Partition2(t *testing.T) {
	// Partition Value = 2
	testGtin14(t, "300800000001004000000001", "00000000001014")
}

func TestGetGtin14Partition3(t *testing.T) {
	// Partition Value = 3
	testGtin14(t, "300C00000010004000000001", "00000000010016")
}

func TestGetGtin14Partition4(t *testing.T) {
	// Partition Value = 4
	testGtin14(t, "301000000080004000000001", "00000000100014")
}

func TestGetGtin14Partition5(t *testing.T) {
	// Partition Value = 5
	testGtin14(t, "301400000400004000000001", "00000001000016")
}

func TestGetGtin14Partition6(t *testing.T) {
	// Partition Value = 6
	testGtin14(t, "301800004000004000000001", "00000010000014")
	testGtin14(t, "30143639F84191AD22901607", "00888446671424")
}

func TestGetGtin14Positive(t *testing.T) {
	// Should allow company prefix = 0
	testGtin14(t, "301800000000004000000001", "00000000000017")
	// Should allow item reference = 0
	testGtin14(t, "301800004000000000000001", "00000010000007")
}

func TestGetGtin14Negative(t *testing.T) {
	testGtin14Error(t, "E2801160600002054CC2096F", "Not a properly encoded SGTIN")
	testGtin14Error(t, "300000181C2CCA93A8B43711", "invalid item reference for SGTIN-96 conversion")
	testGtin14Error(t, "3018000040000040000000011", "Not a properly encoded SGTIN")
	testGtin14Error(t, "30180000400000400000000", "Not a properly encoded SGTIN")
	testGtin14Error(t, "301C00004000004000000001", "invalid partition value")
	testGtin14Error(t, "30244032EACFFD45202001E8", "invalid item reference")
}

func TestGetCompanyPrefixByEpc(t *testing.T) {
	// partition Value = 0
	testCompanyPrefix(t, "300000000000044000000001", 1)
	testCompanyPrefix(t, "300000662D3D311048C6D8D9", 428560204)
	testCompanyPrefix(t, "3000011B896A506B29C18539", 1189239444)
	// Partition Value = 1
	testCompanyPrefix(t, "300400000000204000000001", 1)
	// Partition Value = 2
	testCompanyPrefix(t, "300800000001004000000001", 1)
	// Partition Value = 3
	testCompanyPrefix(t, "300C00000010004000000001", 1)
	// Partition Value = 4
	testCompanyPrefix(t, "301000000080004000000001", 1)
	// Partition Value = 5
	testCompanyPrefix(t, "301800000000004000000001", 0)
	// Partition Value = 6
	testCompanyPrefix(t, "301800004000004000000001", 1)
	testCompanyPrefix(t, "30143639F84191AD22901607", 888446)
}

func TestGetSGTIN96UrnPartition0(t *testing.T) {
	expectedUrn := EPCPureURIPrefix + "000000000001.1.1"
	urn, err := GetSGTINPureURI("300000000000044000000001")
	if err != nil {
		t.Fatalf("[FAIL] expecting proper encode to URN. Error: %s", err.Error())
	}
	if urn != expectedUrn {
		t.Fatalf("[FAIL] expecting %s to be %s", urn, expectedUrn)
	}
}

func TestGetSGTIN96UrnPartition1(t *testing.T) {
	expectedUrn := EPCPureURIPrefix + "00000000001.01.1"
	urn, err := GetSGTINPureURI("300400000000204000000001")
	if err != nil {
		t.Errorf("[FAIL] expecting proper encode to URN")
	}
	if urn != expectedUrn {
		t.Errorf("[FAIL] expecting %s to be %s", urn, expectedUrn)
	}
}

func TestGetSGTIN96UrnPartition2(t *testing.T) {
	expectedUrn := EPCPureURIPrefix + "0000000001.001.1"
	urn, err := GetSGTINPureURI("300800000001004000000001")
	if err != nil {
		t.Errorf("[FAIL] expecting proper encode to URN")
	}
	if urn != expectedUrn {
		t.Errorf("[FAIL] expecting %s to be %s", urn, expectedUrn)
	}
}

func TestGetSGTIN96UrnPartition3(t *testing.T) {
	expectedUrn := EPCPureURIPrefix + "000000001.0001.1"
	urn, err := GetSGTINPureURI("300C00000010004000000001")
	if err != nil {
		t.Errorf("[FAIL] expecting proper encode to URN")
	}
	if urn != expectedUrn {
		t.Errorf("[FAIL] expecting %s to be %s", urn, expectedUrn)
	}
}

func TestGetSGTIN96UrnPartition4(t *testing.T) {
	expectedUrn := EPCPureURIPrefix + "00000001.00001.1"
	urn, err := GetSGTINPureURI("301000000080004000000001")
	if err != nil {
		t.Errorf("[FAIL] expecting proper encode to URN")
	}
	if urn != expectedUrn {
		t.Errorf("[FAIL] expecting %s to be %s", urn, expectedUrn)
	}
}

func TestGetSGTIN96UrnPartition5(t *testing.T) {
	expectedUrn := EPCPureURIPrefix + "0614141.000734.314159"
	urn, err := GetSGTINPureURI("3034257BF400B7800004CB2F")
	if err != nil {
		t.Errorf("[FAIL] expecting proper encode to URN")
	}
	if urn != expectedUrn {
		t.Errorf("[FAIL] expecting %s to be %s", urn, expectedUrn)
	}
}

func TestGetSGTIN96UrnPartition6(t *testing.T) {
	expectedUrn := EPCPureURIPrefix + "000001.0000001.1"
	urn, err := GetSGTINPureURI("301800004000004000000001")
	if err != nil {
		t.Errorf("[FAIL] expecting proper encode to URN")
	}
	if urn != expectedUrn {
		t.Errorf("[FAIL] expecting %s to be %s", urn, expectedUrn)
	}
}
