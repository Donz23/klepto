package anonymiser

import (
	"reflect"
	"strings"

	"github.com/hellofresh/klepto/pkg/config"
	"github.com/hellofresh/klepto/pkg/database"
	"github.com/hellofresh/klepto/pkg/reader"
	log "github.com/sirupsen/logrus"
)

// literalPrefix defines the constant we use to prefix literals
const literalPrefix = "literal:"

// anonymiser anonymises MySQL tables
type anonymiser struct {
	reader.Reader
	tables config.Tables
}

// NewAnonymiser returns an initialised instance of MySQLAnonymiser
func NewAnonymiser(source reader.Reader, tables config.Tables) reader.Reader {
	return &anonymiser{source, tables}
}

func (a *anonymiser) ReadTable(tableName string, rowChan chan<- database.Row, opts reader.ReadTableOpt) error {
	logger := log.WithField("table", tableName)

	logger.Info("Loading anonymiser config")
	table, err := a.tables.FindByName(tableName)
	if err != nil {
		logger.WithError(err).Warn("the table is not configured to be anonymised")
		return a.Reader.ReadTable(tableName, rowChan, opts)
	}

	if len(table.Anonymise) == 0 {
		logger.Debug("Skipping anonymiser")
		return a.Reader.ReadTable(tableName, rowChan, opts)
	}

	// Create read/write chanel
	rawChan := make(chan database.Row)

	// Anonimise the rows
	go func() {
		for {
			row, more := <-rawChan
			if !more {
				close(rowChan)
				return
			}

			for column, fakerType := range table.Anonymise {
				if strings.HasPrefix(fakerType, literalPrefix) {
					row[column] = strings.TrimPrefix(fakerType, literalPrefix)
					continue
				}

				for name, faker := range Functions {
					if fakerType != name {
						continue
					}

					row[column] = faker.Call([]reflect.Value{})[0].String()
				}
			}

			rowChan <- row
		}
	}()

	// Read from the reader
	a.Reader.ReadTable(tableName, rawChan, opts)

	return nil
}
