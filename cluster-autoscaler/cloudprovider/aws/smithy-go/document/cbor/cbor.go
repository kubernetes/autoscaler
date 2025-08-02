// Package cbor implements reflective encoding of Smithy documents for
// CBOR-based protocols.
//
// This package is NOT caller-facing and is not suitable for general
// application use. Callers using the document type with SDK clients should use
// the embedded NewLazyDocument() API in the SDK package to create document
// types.
package cbor
