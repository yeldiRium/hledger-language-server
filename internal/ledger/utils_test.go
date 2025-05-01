package ledger

import (
	"testing"

	participleLexer "github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
)

// TODO: make this more generic. maybe use reflection? so that it doesn't have
// to change every time Journal changes.
func pruneMetadataFromAst(ast *Journal) {
	for _, entry := range ast.Entries {
		switch entry := entry.(type) {
		case *AccountDirective:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		case *RealPosting:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		case *VirtualPosting:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		case *VirtualBalancedPosting:
			entry.AccountName.Pos = participleLexer.Position{}
			entry.AccountName.EndPos = participleLexer.Position{}
		}
	}
}

type testParserConfig struct {
	// parser is the parser that will be tested
	parser *JournalParser

	// expectedAst is the AST that is compared to the parser result
	expectedAst *Journal

	// input is the content that will be parsed
	input string

	// fileName is used for metadata during parsing and will be present in AST
	// metadata
	fileName string

	// assertMetadata determines whether parser metadata (pos, endpos, etc) will
	// be included in the assertion or pruned beforehand. default is false
	assertMetadata bool
}

type testParserOption func(config *testParserConfig)

func ParserInput(input string) testParserOption {
	return func(config *testParserConfig) {
		config.input = input
	}
}

func ParsedFileName(fileName string) testParserOption {
	return func(config *testParserConfig) {
		config.fileName = fileName
	}
}

func AssertMetadata() testParserOption {
	return func(config *testParserConfig) {
		config.assertMetadata = true
	}
}

func ExpectAst(ast *Journal) testParserOption {
	return func(config *testParserConfig) {
		config.expectedAst = ast
	}
}

func AssertParser(t *testing.T, parser *JournalParser, opts ...testParserOption) bool {
	config := testParserConfig{
		parser: parser,
		input:    "",
		fileName: "test.journal",
		assertMetadata: false,
	}

	for _, option := range opts {
		option(&config)
	}

	return runAssertParser(t, config)
}

func runAssertParser(t *testing.T, config testParserConfig) bool {
	parser := config.parser
	ast, err := parser.ParseString(config.fileName, config.input)
	if !assert.NoError(t, err) {
		return false
	}

	if !config.assertMetadata {
		pruneMetadataFromAst(ast)
	}

	return assert.Equal(t, config.expectedAst, ast)
}

// TODO: add and implement method AssertParserFails
func AssertParserFails(t *testing.T, parser *JournalParser, input string) {
	_, err := parser.ParseString("test.journal", input)
	assert.Error(t, err, "expected parser to fail")
}
