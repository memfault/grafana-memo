package main

import "strings"

// CsvStringVar
type CsvStringVar []string

// String
func (c *CsvStringVar) String() string {
	return strings.Join(*c, ",")
}

// Set
func (c *CsvStringVar) Set(value string) error {
	*c = strings.Split(value, ",")
	return nil
}
