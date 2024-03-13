// Package bin identifies the card scheme behind a payment card number from
// its leading digits (the IIN, historically called the BIN) and knows which
// number lengths each scheme issues.
//
// The range table is compiled into the package, so detection is a handful
// of string comparisons with no allocation and no I/O. It covers the public
// ranges of the major international schemes; it is not a substitute for a
// full BIN database when issuer, country or product details are needed.
//
// Detect answers "which scheme does this prefix belong to", Validate adds
// the length check, and Scheme values round-trip through text so they can
// be used directly in flags, JSON and configuration files.
package bin
