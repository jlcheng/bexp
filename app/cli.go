package app

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/scorch"
	"github.com/blevesearch/bleve/search/highlight/highlighter/ansi"
	"github.com/blevesearch/bleve/search/query"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func Cli(searchStr string) error {
	indexDir, err := homedir.Expand("~/tmp/bexp-idx")
	if err != nil {
		return err
	}
	fmt.Println("indexDir:", indexDir, "query:", searchStr)

	index, err := bleve.Open(indexDir)
	if err != nil {
		return err
	}
	defer index.Close()

	//registry.RegisterHighlighter("ansi", ansi.Constructor)
	searchStr = "TODO DEADLINE"
	q := query.NewQueryStringQuery(searchStr)
	sr := bleve.NewSearchRequest(q)
	sr.Fields = []string{"Body"}
	sr.IncludeLocations = true
	sr.Explain = true
	sr.Highlight = bleve.NewHighlightWithStyle(ansi.Name)
	results, err := index.Search(sr)
	if err != nil {
		return err
	}
	for _, hit := range results.Hits {
		fmt.Println("ID:", hit.ID)
		fields := hit.Fields
		body := fields["Body"]
		_ = body

		for field := range hit.Locations {
			for term := range hit.Locations[field] {
				locs := hit.Locations[field][term]
				sidx := make([]string, len(locs))
				for i, loc := range locs {
					sidx[i] = fmt.Sprintf("%v", loc.Start)
				}
				fmt.Printf(" term \"%v\" found at indexes: %v\n", term, strings.Join(sidx, ", "))
			}
		}

		for field := range hit.Fragments {
			fmt.Printf(" fragements for %v: %v of length %v\n", field, reflect.TypeOf(hit.Fragments[field]), len(hit.Fragments[field]))
		}

	}




	return nil
}

type Doc struct {
	ID string
	Body string
}

func Highlight() *bleve.HighlightRequest {
	return bleve.NewHighlight()
}

func Index() error {
	indexDir, err := homedir.Expand("~/tmp/bexp-idx")
	if err != nil {
		return err
	}
	dataDir, err := homedir.Expand("~/org")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("indexDir:", indexDir, "dataDir:", dataDir)

	index, err := bleve.NewUsing(indexDir, bleve.NewIndexMapping(), scorch.Name, scorch.Name, nil)
	if err == bleve.ErrorIndexPathExists {
		if err2 := os.RemoveAll(indexDir); err2 != nil {
			fmt.Println("Cannot delete indexDir", indexDir)
			return err2
		}
		var err2 error
		index, err2 = bleve.NewUsing(indexDir, bleve.NewIndexMapping(), scorch.Name, scorch.Name, nil)
		if err2 != nil {
			fmt.Println("cannot create index", indexDir)
			return err2
		}
	}
	defer index.Close()
	finfos, err := ioutil.ReadDir(dataDir)
	if err != nil {
		return err
	}
	batch := index.NewBatch()
	for _, finfo := range finfos {
		if !finfo.IsDir() {
			body, err := ioutil.ReadFile(filepath.Join(dataDir, finfo.Name()))
			if err != nil {
				return err
			}
			err = batch.Index(finfo.Name(), Doc{ID: finfo.Name(), Body:string(body)})
			if err != nil {
				return err
			}
			fmt.Println("index", finfo.Name())
		}
	}
	err = index.Batch(batch)
	if err != nil {
		return err
	}

	fmt.Println("index complete")
	return nil
}