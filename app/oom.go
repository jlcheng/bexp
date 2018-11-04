package app

import (
	"errors"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/scorch"
	"github.com/blevesearch/bleve/mapping"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	_ "net/http/pprof"
)

const (
	BODY = "Body"
	BATCH_SIZE = 1000
)


func OOMIndex(useScorch bool, dataDir, idxDir string) error {
	if info, _ := os.Stat(idxDir); info != nil {
		return errors.New("idxDir not-empty, will not proceed")
	}

	go func() {
		log.Println("starting net/http/pprof server at localhost:8888")
		log.Println(http.ListenAndServe("localhost:8888", nil))
	}()


	var idx bleve.Index
	var err error
	if useScorch {
		idx, err = bleve.NewUsing(idxDir, NewIndexMapping(), scorch.Name, scorch.Name, nil)
	} else {
		idx, err = bleve.New(idxDir, NewIndexMapping())
	}
	if err != nil {
		return err
	}

	idxHelper := indexHelper{
		index: idx,
		totalSize: 0,
		count: 0,
		batch: idx.NewBatch(),
	}

	dataInfo, err := os.Stat(dataDir)
	err = idxHelper.indexFiles(dataDir, dataInfo)
	if err != nil {
		return err
	}
	return nil
}

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
	totalSize int64
	count int

	batch *bleve.Batch
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
	i.batch.Index(path, doc)
	i.totalSize = i.totalSize + info.Size()
	i.count = i.count + 1

	if i.batch.Size() >= BATCH_SIZE {
		if err := i.index.Batch(i.batch); err != nil {
			return err
		}
		i.batch.Reset()
	}

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
