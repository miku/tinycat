package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/highlight/fragment_formatters/ansi"
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
		defer index.Close()

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
			if len(raw) < 2 {
				continue
			}
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

	fullQuery := fmt.Sprintf("%s %s", *query, strings.Join(flag.Args(), " "))

	index, err := bleve.Open("tc.bleve")
	if err != nil {
		log.Fatal(err)
	}
	q := bleve.NewQueryStringQuery(fullQuery)
	searchRequest := bleve.NewSearchRequest(q)
	searchRequest.Highlight = bleve.NewHighlightWithStyle("ansi")
	ff := ansi.NewFragmentFormatter(ansi.Underscore)
	bleve.Config.Cache.Highlighters["ansi"].SetFragmentFormatter(ff)
	sr, err := index.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sr)
}
