package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
)

const NUM_FIELDS int = 22
const SEVERITY_HEADING string = "severity"

// Index of fields in each CSV format record
const (
	IND_MESSAGE_BODY  = 11
	IND_MSG_TIMESTAMP = 12
	IND_SEVERITY      = 16
	IND_SOURCE_APP    = 17
)

// Processes OMC application log data in CSV format and writes out
// message body and other minimal data. Also, filters output based on
// criteria. Sample header line and a single log entry below.
//"@timestamp","@version","_id","_index","_score","_type",datacenterid,host,instancename,instancetype,logfilename,messagebody,messagetimestamp,oraclebusinessunit,"pod-tag",port,severity,sourceapp,sourceappversion,"tailed_path",targethost,tenantid
//"January 14th 2019, 04:29:49.921",1,QYSfSmgByigPABYRjYAB,"omc-virtualtenant-log-2019.01.14",,doc,dsa,"ops_poll_virtualtenant_logs.1.713kvchfoxwbrr4g2es8djdri.ops_stack","ost-omc-omcnode74cbf37e",Production,"cm.pinlog"," cm  cm:5725.-155481792  cm_child.c(133):4302 2:cm:cm:1:-155481792:0:1547219297:0::::
//	CMAP: Polling for shutdown request","January 14th 2019, 04:29:50.163",cgbu,"cm-7d6d49fdfd-qx8g2","37,671",DEBUG,cmD,18C,"/omc_logs/cm.pinlog",cm,virtualtenant
func main() {
	var source_app string

	// Flag -app gives the source application
	flag.StringVar(&source_app, "app", "",
		"Source application to filter log records by")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [-app sourceapp]\n", path.Base(os.Args[0]))
		fmt.Fprintf(flag.CommandLine.Output(), "  Reads OMC log " +
			"file from stdin, filters metadata and writes raw log " +
			"records to stdout\n")
		fmt.Fprintf(flag.CommandLine.Output(),
			"  -app <sourceapp> Filter records by source " +
			"application, for example, cm\n")
		//flag.PrintDefaults()
	}
	flag.Parse()

	// Read input log records and print relevant fields
	r := csv.NewReader(os.Stdin)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else if len(record) != NUM_FIELDS {
			log.Fatalf("Input record should have %d fields, found %d",
				NUM_FIELDS, len(record))
		}

		// Skip the header line that has column headings
		if record[IND_SEVERITY] != SEVERITY_HEADING {
			// Filter by source app if flag provided
			if source_app == "" ||
				record[IND_SOURCE_APP] == source_app {
				fmt.Printf("%s %s\n", record[IND_SEVERITY],
					record[IND_MSG_TIMESTAMP])
				fmt.Println(record[IND_MESSAGE_BODY])
			}
		}
	}
} // end main
