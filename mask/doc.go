// Package mask prepares card numbers for display and logging.
//
// Normalize turns user-entered input such as "4111 1111 1111 1111" or
// "4111-1111-1111-1111" into a plain digit string and rejects anything that
// is not a digit or a separator. Mask and Last4 build on it to produce the
// truncated forms that may appear in logs, receipts and support tools:
// never more than the first six and last four digits of a number.
//
// The package does not validate the number. Pair it with the luhn and bin
// packages when correctness matters as well as presentation.
package mask
