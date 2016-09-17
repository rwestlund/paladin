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
