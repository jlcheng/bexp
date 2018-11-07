package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/scorch"
	"github.com/blevesearch/bleve/mapping"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var idxDir string
	var useScorch bool
	flag.StringVar(&idxDir, "idxDir", "", "idxDir")
	flag.BoolVar(&useScorch, "scorch", false, "scorch")

	flag.Parse()

	var index bleve.Index
	var err error
	if useScorch {
		index, err = bleve.NewUsing(idxDir, NewIndexMapping(), scorch.Name, scorch.Name, nil)
	} else {
		index, err = bleve.New(idxDir, NewIndexMapping())
	}
	if err != nil {
		log.Fatal(err)
	}

	idxHelper := indexHelper{
		index:     index,
		totalSize: 0,
		count:     0,
		batch:     index.NewBatch(),
		idxStime: time.Now(),
	}


	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		file := scanner.Text()
		if strings.HasPrefix(file, "#") {
			continue
		}
		fmt.Println("index", file)
		doc := SimpleDoc{
			Body: debugReadFile(file),
		}
		err = idxHelper.batch.Index(file, doc)
		if err != nil {
			log.Fatal(err)
		}
	}
	index.Batch(idxHelper.batch)

}

const (
	BODY = "Body"
	BATCH_SIZE = 100
)


type SimpleDoc struct {
	Body string
}

func NewIndexMapping() mapping.IndexMapping {
	imap := bleve.NewIndexMapping()

	// needed because bleve will map SimpleDoc to the "_default" bleve-type
	main_dmap := bleve.NewDocumentMapping()
	imap.AddDocumentMapping("_default", main_dmap)

	// configure the fields in simpleDoc
	body_fmap := bleve.NewTextFieldMapping()
	main_dmap.AddFieldMappingsAt(BODY, body_fmap)

	return imap
}

type indexHelper struct {
	index bleve.Index
	batch *bleve.Batch
	totalSize int64
	count int
	idxStime time.Time
}

func (i *indexHelper) indexFiles(path string, info os.FileInfo) error {
	// Do not index .git
	if info.IsDir() && strings.HasSuffix(path, ".git") {
		return filepath.SkipDir
	}

	// recurse into directory
	if info.IsDir() {
		cinfos, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}
		for _, childInfo := range cinfos {
			childPath := filepath.Join(path, childInfo.Name())
			cerr := i.indexFiles(childPath, childInfo)
			if cerr != nil && cerr != filepath.SkipDir {
				return cerr // bail on legitimate errors
			}
		}
	}

	// File name must contain a '.'
	if strings.LastIndexByte(path, '.') < strings.LastIndexByte(path, '/') {
		return nil
	}


	// Finally, index the heck out of this file
	doc := SimpleDoc{
		Body: debugReadFile(path),
	}
	if i.batch == nil {
		i.batch = i.index.NewBatch()
	}
	fmt.Println("indexing", path)
	i.batch.Index(path, doc)
	i.totalSize = i.totalSize + info.Size()
	i.count = i.count + 1

	if i.batch.Size() >= BATCH_SIZE {
		if err := i.index.Batch(i.batch); err != nil {
			return err
		}
		i.batch.Reset()
		i.batch = i.index.NewBatch()
		i.printProgress()
		fmt.Printf("indexing %v\n", path)
	}

	return nil
}

func (i *indexHelper) printProgress() {
	fmt.Printf("indexed %v files, %v kb, %v seconds elapsed\n", i.count, i.totalSize/1024, time.Since(i.idxStime))

}

func (i *indexHelper) complete() error {
	if i.batch != nil {
		if err := i.index.Batch(i.batch); err != nil {
			return err
		}
	}
	i.printProgress()
	return nil
}


func debugReadFile(fileName string) string {
	f, err := os.Open(fileName)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	defer f.Close()
	s, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return string(s)
}
