package tests

import (
	"fmt"
	"github.com/blevesearch/bleve/registry"
	"reflect"
	"testing"

	"github.com/blevesearch/bleve/analysis"
)

// TestReplace demonstrates the purpose of a TokenFilter: Transforming one TokenStream into another TokenStream.
func TestReplace(t *testing.T) {

	inputTokenStream := analysis.TokenStream{
		&analysis.Token{
			Term: []byte("walk"),
		},
		&analysis.Token{
			Term: []byte("park"),
		},
	}

	expectedTokenStream := analysis.TokenStream{
		&analysis.Token{
			Term: []byte("walk"),
		},
		&analysis.Token{
			Term: []byte("spark"),
		},
	}

	registry.RegisterTokenFilter(ReplaceTokenFilterID, ReplaceTokenFilterCtor)
	cache := registry.NewCache()
	mapping := map[string]string{
		"park": "spark",
	}
	var replaceTokenFilter analysis.TokenFilter
	var err error
	var outputTokenStream analysis.TokenStream
	// The expected way to configure a TokenFilter - by passing the ctor args in a map
	replaceTokenFilter, err = cache.DefineTokenFilter(ReplaceTokenFilterID, ReplaceTokenFilterCtorArgs(mapping))
	if err != nil {
		t.Fatal(err)
	}
	outputTokenStream = replaceTokenFilter.Filter(inputTokenStream)

	if !reflect.DeepEqual(outputTokenStream, expectedTokenStream) {
		t.Errorf("expected %#v got %#v", expectedTokenStream, outputTokenStream)
	}

	// The easy way to configure a TokenFilter, which can never be used in Bleve
	replaceTokenFilter = NewReplaceTokenFilter(mapping)
	if err != nil {
		t.Fatal(err)
	}
	outputTokenStream = replaceTokenFilter.Filter(inputTokenStream)

	if !reflect.DeepEqual(outputTokenStream, expectedTokenStream) {
		t.Errorf("expected %#v got %#v", expectedTokenStream, outputTokenStream)
	}
}

const ReplaceTokenFilterID = "replace_token_filter"
const rtf_ctor_1 = "mapping"
func ReplaceTokenFilterCtorArgs(mapping map[string]string) map[string]interface{} {
	return map[string]interface{} {
		"type":     ReplaceTokenFilterID,
		rtf_ctor_1: mapping,
	}
}
func ReplaceTokenFilterCtor(config map[string]interface{}, _ *registry.Cache) (analysis.TokenFilter, error) {
	mapping, ok := config[rtf_ctor_1].(map[string]string)
	if !ok {
		return nil, fmt.Errorf("must specify '%s'", rtf_ctor_1)
	}
	return NewReplaceTokenFilter(mapping), nil
}

type ReplaceTokenFilter struct {
	mapping map[string]string
}
func NewReplaceTokenFilter(mapping map[string]string) *ReplaceTokenFilter {
	return &ReplaceTokenFilter {
		mapping: mapping,
	}
}
func (f *ReplaceTokenFilter) Filter(input analysis.TokenStream) analysis.TokenStream {
	for i, token := range input {
		replacement, ok := f.mapping[string(token.Term)]
		if ok {
			input[i].Term = []byte(replacement)
		}
	}
	return input
}

