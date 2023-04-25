// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.

package auth

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"os"
	"testing"
)

var (
	testEncryptedPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
Proc-Type: 4,ENCRYPTED
DEK-Info: DES-EDE3-CBC,05B7ACED45203763

bKbv8X2oyfxwp55w3MVKj1bfWnhvQgyqJ/1dER53STao3qRS26epRoBc0BoLtrNj
L+Wfa3NeuEinetDYKRwWGHZqvbs/3PD5OKIXW1y/EAlg1vr6JWX8KxhQ0PzGJOdQ
KPcB2duDtlNJ4awoGEsSp/qYyJLKOKpcz893OWTe3Oi9aQpzuL+kgH6VboCUdwdl
Ub7YyTMFBkGzzjOXV/iSJDaxvVUIZt7CQS/DkBq4IHXX8iFUDzh6L297/BuRp3Q8
hDL4yQacl2F2yCWpUoNNkbPpe6oOmL8JHrxXxo+u0pSJELXx0sjWMn7bSRfgFFIE
k08y4wXZeoxHiQDhHmQI+YTikgqnxEWtDYhHYvWudVQY6Wcf1Fdypa1v4I3gv4S9
QwjDRbRcrnPxMkxWmQEM6xGCwWBj8wmFyIQoEA5MJuQZxWdyptEKVtwwI1TB9etn
SlXPUl125dYYBu2ynmR96nBVEZd6BWl+iFeeZnqxDHABOB0AvpI61vt/6c7tIimC
YciZs74XZH/ERs55p0Ng/G23XNu+UGQQptrr2kyRR5JrS0UGKVjivydIK5Lus4c4
NTaKyEJNMbvSUGY5SLfxyp6HZnlbr4aCDAk62+2ZUotr+sVXplCpuxoSc2Qlw0en
y+plCvd2RdQ/EzIFkpi9V/snIvbMvH3Sp/HqFDG8GehFTRvwpCIVqWC+BZYeaERX
n2P4jODz2M8Ns7txv1nB4CyxWgu19398Zit0K0QmG24kCJtLg9spEOmKtoIuVTnU
9ydxmHQjNNtyH+RceZFn07IkWvPveo2BXpK4K9DXE39Z/g1nQzwTqgN8diXxwRuN
Ge97lBWup4vP1TV8nyHW2AppgFVuPynO+XWfZUuCUzxNseB+XOyeqitoM4uvSNax
DQmokjIf4qXC/46EnJ/fd9Ydz4GVQ4TYyxwNCBJK39RdUOcUtyI+A3IbZ+vt2HIV
eiIN2BhdnwbvNTbPs9nc9McM2NtACqDGQsIzRdXcQ8SFDP2DnTVjGu5E8H9dnVrd
FcuUnA9TIbfBkRHOS7yoDHOo4j28g6xePDV5tK0L5C2yyDh+bwWnO5AIg/gdpnuH
wxIZUxFwkD4GvOVtj5Y4W5L+Uy3c94stMPbHE+zGN75DdQRy5aVbDjWqXRB9AEQN
+NSb526oqhv0JyYlZmCqz2ydBxkT4FsShZv/34pkRr3qL5FSTAQTXQAZdiQQbMTe
H3zKyu4GbEUV9WsyriqSq27ptMwFfIqN1NdsWeVWN1mXf2KZDn61EgleeQXmdSZu
XM4Z1n98xjYDwdCkF738j+oRAlSUThBeU/hYbH6Ysff6ON9MPBAAKy3ZxM5tF86e
l0x20lpND2QLLDZbsg/LrCrE6ZzpWkXn4w4PG4lWMAqph0BebSkFqXvUvuds3c39
yptNH3FsyqeyM9kDwbDpBQAvpsDIQJfwAbQPLAiQJhpbixZyG9lqhkKOhYTZhU3l
ufFtnLEj/5G9a8A//MFrXsXePUeBDEzjtEcjPGNxe0ZkuOgYx11Zc0R4oLI7LoHO
07vtw4qCH4hztCJ5+JOUac6sGcILFRc4vSQQ15Cg5QEdBiSbQ/yo1P0hbNtSvnwO
-----END RSA PRIVATE KEY-----`
	testKeyPassphrase = "goisfun"

	// Generated key duplicates the Java SDK test resources

	testPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA8Wf2KWrO/KEqeKKLN7chonrvVsSOEJcZQykYnCePswsFYNDB
/6O7njPx12Y0tRB8vaLVFNuoylGPLTNCfCy5veeVZnBX5/MENSh9/a3KMvSaKG8P
o5eNqYdyuZTaHUsa2vOjcyQfNZv9p1G4d/HhiDmqGP8KggV5ZhD0j7UteUt5OZSL
e3oGZ3M0/P5VtXZfD1Fq734va2ta4Q+zxE5IHJn7GW/VwK7TqctyuJFznbapjAtQ
pUsJrig6jqnyMd1rIs4hS8zimxfVo05mbr7Z1r38XIeXYiqICCry8N21BICAItrE
L25SOkzcA/sB1PvcSFIAV1i1hLKrnBU/eg+T+wIDAQABAoIBAQCI7GKfE0nb2L3y
Np+oNmMJeZkPKeU6W7mkckbXK0lCUFn4k++1Q/VCwkvF1N7IZFWcaiNZ9U1DlAcV
qCFptSSVJimDNO1nTltwm0r6+/vX8w0NKhFAxNFA+uaDhH5CZzsQPWjUAgUBrzys
DpoGzlcRoUNtchtPrDMzRSKx8B2e0aooD9kHQrXPYftaUi9rW+C44GXqt00ulZtZ
gyXbIy8WB0ZlQt40aS+1NyNcZyBb2bf0ldwZbRwh4ZzdWyoD/v2Vsi0RL0aaV1T1
GtnnFIRJH6WeZTjoBzCw88VAwZdRCKxZaJ0mB08VS+bAIBEXN6/Ts3SKUQ87Y+I0
xgAuh1/BAoGBAPq/xQTEv4YB14CYEPo2CIV38poxF5GonlBmEq5kedFWNLfO3+ge
KsHYPuB1ZUc89a0zUnaq7FdYSJGD2xmLNqN+R4OVTxPztpakglKHWhisK+2ocBL7
lsJkBTg9El+N4a2MQohTDfXf/br/2HwZeCN5G2FvixUzUm1pcYZpfTuZAoGBAPZ2
Gv/KrL93btabNoMLprw4Z/ElVoBo5NQpn4BNuVeABc60v+RwWxsAvwHOTGYY5TZR
hQeaDhvusiH7vT9ozNb7YpehFaay2pkgytFAvcr2QhFcjbOGq2cm4puuGZGN3uOL
ErhwxbNmKVxhjpVUUNkZdYmufkA6i4ltAvLNByizAoGAYrZsEVyDKXZAKFe1F0t+
P0zhLOJ2rNj8uhn08MKNUmPljRbb/r0hh/5hgmu02z6cWPsDU8QmFpyitOZ7sqqj
b+meraZx4yDmmJda1rKCPYRKJt1QgaiZyR0nEOS5/vQUDAZTiudnb4wmjx95UiGU
siJTLSCEWGxD3t7L2mZc7sECgYBsec0mWmEwIHQTVttmUEGBxF3TYHizKffVfcBr
K0pxPbLQqPNwqxceSnTHabJsmXaBMt4XW3HsT2Ht3SwNdaX61UgursKlzUCzdyBt
e05Nv5eSpqbjpllYnF/O35D3ZHb+tZ52uYP6kvOPaozkIuk2tKLsB3Yf9OSnhuhu
T1lgSwKBgDLm/qguyd2gdhOBhVXJkfLRKZwogQI1ZdZYjNssCVFmnX+FvEJV+4cS
BjYWARWBrZbkAz3+8y9EIHV9aRmnX1jGOVUJ0MhaYyZ18KTzi32gAG8s7V7r57Sv
1D6imvCUmX3q5d1QuikaFw0VBX6S/FPFLneJXybW+GHwsgcGToxU
-----END RSA PRIVATE KEY-----`

	// Generated RPST duplicates the Java SDK test resources
	rpst = `eyJraWQiOiJhc3ciLCJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJvY2lkMS5kYXRhd2FyZWhvdXNlLmN1c3RvbWVyLWJpZy1kYi0xIiwiaXNzIjoiYXV0aFNlcnZpY2Uub3JhY2xlLmNvbSIsInJlc190ZW5hbnQiOiJjdXN0b21lci10ZW5hbnQtMSIsInB0eXBlIjoicmVzb3VyY2UiLCJyZXNfdHlwZSI6ImRhdGF3YXJlaG91c2UiLCJhdWQiOiJvY2kiLCJvcGMtdGFnIjoiVjEsb2NpZDEuZHluYW1pY2dyb3VwLmRldi4uYWFhYWFhYWEsb28ycHA2M3Y3NTMyNXpiNnJuNHYzYWgzYnRlam5ubTJvdDN1NWJyZnd1Y2VvczRpZDZ0YSIsIm9wYy1idW1wIjoiSDRzSUFBQUFBQUFBQUUyUHpRcUNRQlNGTTExTVJSQXRLbHBHYlFRSE1sSm8xVVAwQXVQTWJSeHJISnVmMHA0LUF4SFA5bjczTzV6bFRWSEJqcGcxSlpHQ2NxMWNoUm04TVNaZDVGM3cwMHZGejhaQnduaVpTZjA0eDBWaThuZGVrSlJMY0o4NnF3cWFmRTJkMGx3bE5RbFQ1Qy04N1ZxRFVVNVR3RlRKaW1ncm9iUllzTjJLT21PVkJCME5EdEh4T2dvbktGaE1ROThiLS1FZWpWdkh2SGZZcG9MZG5CRkxQa1JEcnB5Qjl1R0N2SmFhOVZSclAzU1RoaVR1R3pQQkk1Yjl5LUlBZVp2Z0IxX2paUVFDQVFBQSIsInR0eXBlIjoicmVzIiwicmVzX2lkIjoib2NpZDEuZGF0YXdhcmVob3VzZS5jdXN0b21lci1iaWctZGItMSIsIm9wYy1pbnN0YW5jZSI6Im9jaWQxLmluc3RhbmNlLmpveWZ1bC5kYm5vZGUuMSIsInJlc190YWciOiIiLCJyZXNfY29tcGFydG1lbnQiOiJjdXN0b21lci1jb21wYXJ0bWVudC0xIiwiZXhwIjoxNTUyNjgxMTY4LCJvcGMtY29tcGFydG1lbnQiOiJkYmFhcy1jb21wYXJ0bWVudC0xIiwiaWF0IjoxNTUyNjc5OTY4LCJqdGkiOiIzZDQxNDNlOC01ZTMyLTRlMTItYTM4Yy01OTc0NjUwMTA3MDMiLCJ0ZW5hbnQiOiJjdXN0b21lci10ZW5hbnQtMSIsImp3ayI6IntcImtpZFwiOlwiY3VzdG9tZXItZGJub2RlLTFcIixcIm5cIjpcImxzLUFDNGhpS0stMTFVdTFEZ3VLTFE1VGFhZGpNR1hCcDRhMFVFS2w0dnJjcmF3b2V6X3BuUS1pNS1nNV9XTU5xVXdrdUtBcXVTZnlVS25yZEhhV3d4b2RWcmRleTk1T3R4ckIySzNRdzgzaURkcUltSkhfWFp1cERfRHR0SzduS3N6Qy01TFI1Ums3SHF5Y094eEZVNzBNcGduQW9IaVNUM2V0VjJVZlJkNXRtb0dOaTdOSURORWJnSVpmcnczYUVYbHBzaGM2ckpVdUEyOG55ZUNjOVFtOHllMHUwN0UzamlCYmp5RjNFVWhTelNxblFsUlVNVEdaR1ZSZGpfRG9tcEhUVkFPNEJqUnVIZURWWGtWNjh1TzNrSEdTZUVPc2xsZmJZTkpaYUtCQTB3aUxrZkViWVBEWDFwbTM4UFAzcnJFbWhxeElObzdoYVFWakRXRDNKUVwiLFwiZVwiOlwiQVFBQlwiLFwia3R5XCI6XCJSU0FcIixcImFsZ1wiOlwiUlMyNTZcIixcInVzZVwiOlwic2lnXCJ9Iiwib3BjLXRlbmFudCI6ImRiYWFzLXRlbmFudCJ9.LqRt9JXSdcLahdwACjw_p_KHQhKde-NaVZG3zMjzWX6bVad-SRZYWKQSlk6Tq4f1ZNN0uxlP-d2snQAp3Kw-cQRrdCDOmD_0CDgR-yre-YbJDsJbEncczUIbe-ASeq_Sh9zDROVuD_7NdrmUCiVH2g-UYpYkuKKqu_tVjL2uy77W5_DGobPArEFvZ2GnyHT7gVVv12RnINtgr2jJULhegPBfvnp9-fhhZ7_PcsJ7Z5FkPzLtLOwEm3Lbm3veyUVUviu1CSjXnK67KzjS18TVGi723bkxYBf9lYDHfaXh9EEHzPtxeLAl3VrGjwZUv_ih0FRmoM7wgq8HMRjNACMo6g`
)

var ppEnvVars = map[string]string{
	ResourcePrincipalVersionEnvVar:              ResourcePrincipalVersion2_2,
	ResourcePrincipalRegionEnvVar:               string(common.RegionPHX),
	ResourcePrincipalRPSTEnvVar:                 rpst,
	ResourcePrincipalPrivatePEMEnvVar:           testEncryptedPrivateKey,
	ResourcePrincipalPrivatePEMPassphraseEnvVar: testKeyPassphrase,
}

var envVars = map[string]string{
	ResourcePrincipalVersionEnvVar:    ResourcePrincipalVersion2_2,
	ResourcePrincipalRegionEnvVar:     string(common.RegionPHX),
	ResourcePrincipalRPSTEnvVar:       rpst,
	ResourcePrincipalPrivatePEMEnvVar: testPrivateKey,
}

func writeTempFile(data string) (filename string) {
	f, _ := ioutil.TempFile("", "auth-gosdkTest")
	f.WriteString(data)
	filename = f.Name()
	return
}

func removeFile(files ...string) {
	for _, f := range files {
		os.Remove(f)
	}
}

func unsetAllVars() {
	for k := range ppEnvVars {
		os.Unsetenv(k)
	}
}

func setupResourcePrincipalsEnvsWithValues(from map[string]string, enabledVars ...string) {
	if len(enabledVars) == 0 {
		for k, v := range from {
			os.Setenv(k, v)
		}
		return
	}

	for _, v := range enabledVars {
		os.Setenv(v, from[v])
	}
}

func setupResourcePrincipalsEnvsWithPaths(enabledVars ...string) (tempFiles []string) {
	tempFiles = make([]string, 0)
	for _, v := range enabledVars {
		f := writeTempFile(envVars[v])
		tempFiles = append(tempFiles, f)
		os.Setenv(v, f)
	}
	return
}

func TestResourcePrincipalKeyProvider(t *testing.T) {
	unsetAllVars()
	setupResourcePrincipalsEnvsWithValues(ppEnvVars)
	provider, e := ResourcePrincipalConfigurationProvider()

	assert.NoError(t, e)
	assert.NotNil(t, provider)
}

func TestResourcePrincipalKeyProvider_MissingEnvVars(t *testing.T) {
	var testVars = [][]string{
		{ResourcePrincipalVersionEnvVar, ResourcePrincipalRegionEnvVar, ResourcePrincipalPrivatePEMEnvVar,
			ResourcePrincipalPrivatePEMPassphraseEnvVar},
		{ResourcePrincipalVersionEnvVar, ResourcePrincipalRegionEnvVar, ResourcePrincipalPrivatePEMEnvVar},
		{ResourcePrincipalVersionEnvVar, ResourcePrincipalRegionEnvVar},
		{ResourcePrincipalVersionEnvVar},
	}

	for _, v := range testVars {
		unsetAllVars()
		setupResourcePrincipalsEnvsWithValues(ppEnvVars, v...)
		_, e := ResourcePrincipalConfigurationProvider()
		assert.Error(t, e, "should have failed with %s", v)
	}
}

func TestResourcePrincipalKeyProvider_DifferentEnvVarContent(t *testing.T) {
	unsetAllVars()
	setupResourcePrincipalsEnvsWithValues(ppEnvVars)
	tempFiles := setupResourcePrincipalsEnvsWithPaths(ResourcePrincipalRPSTEnvVar)
	defer removeFile(tempFiles...)

	_, e := ResourcePrincipalConfigurationProvider()
	assert.NoError(t, e, "should have not failed")
}

func TestResourcePrincipalConfigurationProvider(t *testing.T) {
	unsetAllVars()
	// Set up the environment as an example consumer (eg, a function) may have it - this injects no passphrase
	setupResourcePrincipalsEnvsWithValues(envVars, ResourcePrincipalVersionEnvVar, ResourcePrincipalRegionEnvVar)
	tempFiles := setupResourcePrincipalsEnvsWithPaths(ResourcePrincipalRPSTEnvVar, ResourcePrincipalPrivatePEMEnvVar)
	defer removeFile(tempFiles...)

	provider, e := ResourcePrincipalConfigurationProvider()
	assert.NoError(t, e)

	tenancyOCID, e := provider.TenancyOCID()
	assert.NoError(t, e)
	assert.Equal(t, "customer-tenant-1", tenancyOCID)

	keyFingerprint, e := provider.KeyFingerprint()
	assert.NoError(t, e)
	assert.Equal(t, "", keyFingerprint)

	userOCID, e := provider.UserOCID()
	assert.NoError(t, e)
	assert.Equal(t, "", userOCID)

	region, e := provider.Region()
	assert.NoError(t, e)
	assert.Equal(t, string(common.RegionPHX), region)

	keyID, e := provider.KeyID()
	assert.NoError(t, e)
	assert.Equal(t, "ST$"+rpst, keyID)

	privateKey, e := provider.PrivateRSAKey()
	assert.NoError(t, e)
	assert.NotNil(t, privateKey)

}
