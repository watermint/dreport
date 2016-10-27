package publisher

import (
	"encoding/csv"
	"os"

	"github.com/cihub/seelog"
)

type CsvPublisher struct {
	OutputFile string

	outFile *os.File
	outCsv  *csv.Writer
}

func (c *CsvPublisher) Headers(headers []string) error {
	return c.outCsv.Write(headers)
}

func (c *CsvPublisher) Row(data []string) error {
	return c.outCsv.Write(data)
}

func (c *CsvPublisher) Open() error {
	out, err := os.Create(c.OutputFile)
	if err != nil {
		seelog.Errorf("Unable to create file: '%s'", c.OutputFile)
		return err
	}
	c.outFile = out
	c.outCsv = csv.NewWriter(out)

	return nil
}

func (c *CsvPublisher) Close() {
	if c.outCsv != nil {
		c.outCsv.Flush()
		c.outCsv = nil
	}
	if c.outFile != nil {
		c.outFile.Close()
		c.outFile = nil
	}
}