package testing

import (
	"fmt"
	"testing"

	participleLexer "github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"

	"github.com/yeldiRium/hledger-language-server/internal/lexing"
)

type MiniToken struct {
	Type  string
	Value string
}

func MakeTokens(lexer *lexing.Lexer, fileName string, miniTokens []MiniToken) []participleLexer.Token {
	tokens := make([]participleLexer.Token, len(miniTokens))
	pos := participleLexer.Position{Filename: fileName, Offset: 0, Line: 1, Column: 1}
	for i, token := range miniTokens {
		tokens[i] = participleLexer.Token{
			Type:  lexer.Symbol(token.Type),
			Value: token.Value,
			Pos:   pos,
		}
		pos.Advance(token.Value)
	}
	return tokens
}

// NullTokenPositions sets offset, line and column of all tokens to zero to make
// comparison of large amounts of tokens more consistent, although less precise.
func NullTokenPositions(tokens []participleLexer.Token) []participleLexer.Token {
	nulledTokens := make([]participleLexer.Token, len(tokens))

	for i, token := range tokens {
		nulledToken := token
		nulledToken.Pos = participleLexer.Position{
			Filename: token.Pos.Filename,
		}
		nulledTokens[i] = nulledToken
	}

	return nulledTokens
}

func RunLexerWithFileName(lexerDefinition *lexing.LexerDefinition, input string, fileName string) (*lexing.Lexer, []participleLexer.Token, error) {
	lexer, err := lexerDefinition.LexString(fileName, input)
	if err != nil {
		return nil, nil, err
	}
	tokens := make([]participleLexer.Token, 0)
	for token, err := lexer.Next(); token.Type != participleLexer.EOF; token, err = lexer.Next() {
		if err != nil {
			return nil, nil, err
		}
		tokens = append(tokens, token)
	}

	fmt.Printf("tokens: %#v\n", tokens)

	return lexer, tokens, nil
}

type testLexerConfig struct {
	// lexerDefinition is the lexer that will be tested
	lexerDefinition *lexing.LexerDefinition

	// input is the content that will be lexed
	input string

	// fileName is used for metadata during lexing and will be present in token
	// metadata
	fileName string

	// expectedTokens are the participle tokens that the lexers output will be
	// compared to. You probably want to use expectedMiniTokens instead.
	expectedTokens []participleLexer.Token

	// expectedMiniTokens are the mini tokens that the lexers output will be
	// compared to. They are first converted to participle tokens using the
	// instantiated lexer.
	expectedMiniTokens []MiniToken

	// includedSymbolNames are symbols that are included when collecting the lexed
	// tokens. If it is not set, all tokens are included.
	includedSymbolNames []string

	// excludedSymbolNames are symbols that are excluded when collecting the lexed
	// tokens. They will not be included when comparing with the expected (mini)
	// tokens.
	excludedSymbolNames []string

	// If onlyIncludeExplicitlyExpectedSymbols is set to true, only symbol names
	// that are used in the expected tokens are collected and compared. This is
	// set to true by default.
	onlyIncludeExplicitlyExpectedSymbols bool

	// ignoreTokenPositions determines whether the positions of lexed tokens are
	// compared when asserting the expected tokens. Is set to true by default,
	// because excluding symbols that are not relevant to a test is a common
	// practice. Disable this to explicitly compare token positions.
	ignoreTokenPositions bool
}

type testLexerOption func(config *testLexerConfig)

func LexerInput(input string) testLexerOption {
	return func(config *testLexerConfig) {
		config.input = input
	}
}

func LexedFileName(fileName string) testLexerOption {
	return func(config *testLexerConfig) {
		config.fileName = fileName
	}
}

func ExpectTokens(expectedTokens []participleLexer.Token) testLexerOption {
	return func(config *testLexerConfig) {
		config.expectedTokens = expectedTokens
	}
}

func ExpectMiniTokens(expectedMiniTokens []MiniToken) testLexerOption {
	return func(config *testLexerConfig) {
		config.expectedMiniTokens = expectedMiniTokens
	}
}

func IncludeSymbols(symbolNames []string) testLexerOption {
	return func(config *testLexerConfig) {
		config.includedSymbolNames = symbolNames
	}
}

func ExcludeSymbols(symbolNames []string) testLexerOption {
	return func(config *testLexerConfig) {
		config.excludedSymbolNames = symbolNames
	}
}

func IncludeUnexpectedSymbols() testLexerOption {
	return func(config *testLexerConfig) {
		config.onlyIncludeExplicitlyExpectedSymbols = false
	}
}

func CompareTokenPositions() testLexerOption {
	return func(config *testLexerConfig) {
		config.ignoreTokenPositions = false
	}
}

func AssertLexer(t *testing.T, lexerDefinition *lexing.LexerDefinition, opts ...testLexerOption) bool {
	config := testLexerConfig{
		lexerDefinition:                      lexerDefinition,
		input:                                "",
		fileName:                             "test.journal",
		ignoreTokenPositions:                 true,
		onlyIncludeExplicitlyExpectedSymbols: true,
	}

	for _, option := range opts {
		option(&config)
	}

	if len(config.includedSymbolNames) > 0 && config.onlyIncludeExplicitlyExpectedSymbols {
		panic("can not combine includeSymbolNames and onlyIncludeExplicitlyExpectedSymbols. to explicitly include symbols, use IncludeUnexpectedSymbols")
	}

	return runAssertLexer(t, config)
}

func runAssertLexer(t *testing.T, config testLexerConfig) bool {
	lexer, err := config.lexerDefinition.LexString(config.fileName, config.input)
	if !assert.NoError(t, err, "failed to initialize the lexer") {
		return false
	}

	expectedTokens := config.expectedTokens
	if len(config.expectedMiniTokens) > 0 {
		expectedTokens = MakeTokens(lexer, config.fileName, config.expectedMiniTokens)
	}

	includedSymbols := make(map[participleLexer.TokenType]struct{})
	if config.onlyIncludeExplicitlyExpectedSymbols {
		for _, token := range expectedTokens {
			includedSymbols[token.Type] = struct{}{}
		}
	} else {
		for _, includedSymbol := range config.includedSymbolNames {
			includedSymbols[lexer.Symbol(includedSymbol)] = struct{}{}
		}
	}

	excludedSymbols := make(map[participleLexer.TokenType]struct{}, len(config.excludedSymbolNames))
	for _, excludedSymbol := range config.excludedSymbolNames {
		excludedSymbols[lexer.Symbol(excludedSymbol)] = struct{}{}
	}

	tokens, err := lexing.CollectLexerTokensWithIncludeExclude(
		lexer,
		includedSymbols,
		excludedSymbols,
	)
	if !assert.NoError(t, err, "encountered an error while collecting") {
		return false
	}

	fmt.Printf("found tokens: %#v\n", tokens)

	if config.ignoreTokenPositions {
		tokens = NullTokenPositions(tokens)
		expectedTokens = NullTokenPositions(expectedTokens)
	}

	// It is possible not to expect anything, but to just assert that lexing is
	// successful.
	if len(expectedTokens) > 0 {
		if !assert.Equal(t, expectedTokens, tokens) {
			return false
		}
	}

	return true
}

func AssertLexerFails(t *testing.T, lexerDefinition *lexing.LexerDefinition, input string) {
	err := runLexer(lexerDefinition, input)
	assert.Error(t, err, "expected lexer to fail")
}

func runLexer(lexerDefinition *lexing.LexerDefinition, input string) error {
	lexer, err := lexerDefinition.LexString("test.journal", input)
	if err != nil {
		return err
	}
	for token, err := lexer.Next(); token.Type != participleLexer.EOF; token, err = lexer.Next() {
		if err != nil {
			return err
		}
	}

	return nil
}
