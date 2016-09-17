/*
 * Copyright (c) 2016, Randy Westlund. All rights reserved.
 * This code is under the BSD-2-Clause license.
 *
 * This is the main file. Run it to launch the application.
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
