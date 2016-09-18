/*
 * Copyright (c) 2016, Randy Westlund. All rights reserved.
 * This code is under the BSD-2-Clause license.
 *
 * This file contains general utility functions.
 */
package main

/* Remove empty strings from a slice of strings. Returns a new slice. */
func squeeze(string_slice []string) []string {
	var ss []string
	for _, s := range string_slice {
		if s != "" {
			ss = append(ss, s)
		}
	}
	return ss
}

// Return true if the first arg slice contains the second arg.
func contains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}
