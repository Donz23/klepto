package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/hellofresh/klepto/database"
	"github.com/spf13/cobra"
)

func RunSteal(cmd *cobra.Command, args []string) {
	inputConn, err := database.Connect(fromDSN)
	if err != nil {
		color.Red("Error connecting to input database: %s", err.Error())
		return
	}
	defer inputConn.Close()

	dumper := database.NewMySQLDumper(inputConn)
	structure, err := dumper.DumpStructure()
	if err != nil {
		color.Red("Error connecting to input database: %s", err.Error())
		return
	}
	cmd.Print(structure)

	out := make(chan *database.Cell, 100)
	done := make(chan bool)
	go func() {
		for {
			cell, more := <-out
			if more {
				fmt.Println(cell.Value)
			} else {
				done <- true
				return
			}
		}
	}()

	columns, err := dumper.GetColumns("users")
	fmt.Printf("INSERT INTO `users` (%s) VALUES\n", strings.Join(columns, ", "))

	anonymiser := database.NewMySQLAnonymiser(inputConn)
	err = anonymiser.DumpTable("users", out)
	if err != nil {
		color.Red("Error stealing data: %s", err.Error())
		return
	}
	close(out)
	<-done

	// outputConn, err := dbConnect(*toDSN)
	// if err != nil {
	// 	return err
	// }
}
