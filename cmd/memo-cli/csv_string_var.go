package main

import "strings"

type CsvStringVar []string

func (c *CsvStringVar) String() string {
	return strings.Join(*c, ",")
}

func (c *CsvStringVar) Set(value string) error {
	*c = strings.Split(value, ",")
	return nil
}
