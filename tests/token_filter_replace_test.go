package tests

import (
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

	replaceFilter := ReplaceTokenFilter{}
	replaceFilter["park"] = "spark"
	ouputTokenStream := replaceFilter.Filter(inputTokenStream)

	if !reflect.DeepEqual(ouputTokenStream, expectedTokenStream) {
		t.Errorf("expected %#v got %#v", expectedTokenStream, ouputTokenStream)
	}
}

type ReplaceTokenFilter map[string]string
func (f *ReplaceTokenFilter) Filter(input analysis.TokenStream) analysis.TokenStream {
	for i, token := range input {
		replacement, ok := (*f)[string(token.Term)]
		if ok {
			input[i].Term = []byte(replacement)
		}
	}
	return input
}

