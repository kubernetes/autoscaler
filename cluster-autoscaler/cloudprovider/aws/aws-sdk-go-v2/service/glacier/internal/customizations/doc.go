// Package customizations provides customizations for the Glacier API client.
//
// # Computing tree hash and sha256 checksum
//
// Glacier requires not only a sha256 checksum header, but also a tree hash. These
// can be set as inputs to the relevant commands, but in most cases this would
// just be tedious boilerplate. So if the checksums have not been provided, the
// client will automatically calculate them.
package customizations
