package main

import "strings"

// CsvStringVar
type CsvStringVar []string

// String
func (c *CsvStringVar) String() string {
	ct := *c

	// trim the extra spaces
	temp := []string{}
	for _, cv := range ct {
		temp = append(temp, strings.TrimSpace(cv))
	}

	return strings.Join(temp, ",")
}

// Set
func (c *CsvStringVar) Set(value string) error {
	*c = strings.Split(value, ",")
	return nil
}
