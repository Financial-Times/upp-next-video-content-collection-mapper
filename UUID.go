package main

import (
	"crypto/md5"
	"strconv"
	"strings"
)

// UUID is modeled after java.util.UUID
// Holds the most significant and least significant bits of a UUID
type UUID struct {
	msb uint64
	lsb uint64
}

// NewNameUUIDFromBytes creates a type 3 - name based - UUID
// based on the specified byte array.
// Corresponds to java.util.UUID#nameUUIDFromBytes
func NewNameUUIDFromBytes(bytes []byte) *UUID {
	md5Hash := md5.Sum(bytes)
	md5Hash[6] &= 0x0f /* clear version        */
	md5Hash[6] |= 0x30 /* set to version 3     */
	md5Hash[8] &= 0x3f /* clear variant        */
	md5Hash[8] |= 0x80 /* set to IETF variant  */

	var msb uint64
	var lsb uint64

	for i := 0; i < 8; i++ {
		msb = (msb << 8) | (uint64(md5Hash[i]) & 0xff)
	}
	for i := 8; i < 16; i++ {
		lsb = (lsb << 8) | (uint64(md5Hash[i]) & 0xff)
	}

	return &UUID{msb, lsb}
}

// String returns a string representing this UUID.
// Corresponds to java.util.UUID#toString
func (uuid *UUID) String() string {
	parts := make([]string, 5)
	parts[0] = digits(uuid.msb>>32, 8)
	parts[1] = digits(uuid.msb>>16, 4)
	parts[2] = digits(uuid.msb, 4)
	parts[3] = digits(uuid.lsb>>48, 4)
	parts[4] = digits(uuid.lsb, 12)

	return strings.Join(parts, "-")
}

// Corresponds to java.util.UUID#digits
func digits(val uint64, digits uint) string {
	hi := 1 << (digits * 4)
	result := uint64(hi) | (val & (uint64(hi) - uint64(1)))
	return strconv.FormatInt(int64(result), 16)[1:]
}
