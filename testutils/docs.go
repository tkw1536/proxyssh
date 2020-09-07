// Package testutils contains functions used as utility code for tests.
//
// Because of that most of these functions do not have return values of type error, but instead call panic() when something goes wrong.
//
// While these functions are stable and tested themselves, they are not intended to be used directly by proxyssh-external code.
// As such their signatures and edge case behaviour may change without notice.
package testutils
