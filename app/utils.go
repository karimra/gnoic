package app

import (
	"fmt"

	"github.com/olekukonko/tablewriter"
)

func sInListNotEmpty(s string, l []string) bool {
	if len(l) == 0 {
		return true
	}
	for i := range l {
		if s == l[i] {
			return true
		}
	}
	return false
}

func (a *App) handleErrs(errs []error) error {
	numErrors := len(errs)
	if numErrors > 0 {
		for _, e := range errs {
			a.Logger.Debug(e)
		}
		return fmt.Errorf("there was %d error(s)", numErrors)
	}
	return nil
}

func formatTable(table *tablewriter.Table) {
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetAutoMergeCellsByColumnIndex([]int{0})
}
