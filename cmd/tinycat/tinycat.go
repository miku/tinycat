package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
)
import (
	"github.com/blevesearch/bleve"
	"github.com/miku/marc22"
	"github.com/miku/marctools"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	inputFile := flag.String("input", "", "path to the MARC file to index")
	query := flag.String("q", "", "simple query")
	batchSize := flag.Int("size", 10000, "number of records to commit at once")

	flag.Parse()

	if *inputFile != "" {
		file, err := os.Open(*inputFile)

		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		tags := []string{"001", "245.a"}
		fillNA := "NA"
		separator := ""
		skipIncompleteLines := true

		mapping := bleve.NewIndexMapping()
		index, err := bleve.New("tc.bleve", mapping)
		if err != nil {
			log.Fatal(err)
		}

		i := 0
		batch := bleve.NewBatch()

		for {
			if i%*batchSize == 0 {
				err := index.Batch(batch)
				if err != nil {
					log.Fatal(err)
				}
				batch = bleve.NewBatch()
			}
			record, err := marc22.ReadRecord(file)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			line := marctools.RecordToTSV(record, &tags, &fillNA, &separator, &skipIncompleteLines)
			raw := strings.Split(*line, "\t")
			doc := struct {
				Id    string
				Title string
			}{
				Id:    strings.TrimSpace(raw[0]),
				Title: strings.TrimSpace(raw[1]),
			}
			batch.Index(doc.Id, doc)
			i++
		}
		err = index.Batch(batch)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *query == "" {
		return
	}

	index, _ := bleve.Open("tc.bleve")
	q := bleve.NewQueryStringQuery(*query)
	searchRequest := bleve.NewSearchRequest(q)
	searchResult, _ := index.Search(searchRequest)
	fmt.Println(searchResult.String())
}
