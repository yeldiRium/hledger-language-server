package ledger_test

import (
	"encoding/json"
	"fmt"
	"testing"

	participleLexer "github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"

	"github.com/yeldiRium/hledger-language-server/ledger"
)

func TestJournalParser(t *testing.T) {
	filename := "testFile"

	runParser := func(testFileContent string) ([]participleLexer.Token, *ledger.Journal, error) {
		lexer := ledger.NewJournalLexer()
		lex, err := lexer.LexString(filename, testFileContent)
		tokens := make([]participleLexer.Token, 0)
		token, err := lex.Next()
		for err == nil && token.Type != participleLexer.EOF {
			tokens = append(tokens, token)
			token, err = lex.Next()
		}
		fmt.Printf("Tokens: %#v\n", tokens)

		parser := ledger.NewJournalParser()
		ast, err := parser.ParseString(filename, testFileContent)

		jsonAst, _ := json.Marshal(ast)
		fmt.Printf("AST: %s\n", jsonAst)

		return tokens, ast, err
	}

	t.Run("General format", func(t *testing.T) {
		t.Run("Parses a file containing only newlines.", func(t *testing.T) {
			_, ast, err := runParser("\n\n\n")
			assert.NoError(t, err)
			assert.Equal(t, &ledger.Journal{}, ast)
		})
	})

	t.Run("Include directive", func(t *testing.T) {
		t.Run("Parses a file containing an include directive with a path.", func(t *testing.T) {
			_, ast, err := runParser("include some/path.journal\n")
			pruneMetadataFromAst(ast)

			assert.NoError(t, err)
			assert.Equal(t, &ledger.Journal{
				Entries: []ledger.Entry{
					&ledger.IncludeDirective{
						IncludePath: "some/path.journal",
					},
				},
			}, ast)
		})

		t.Run("Allows multiple spaces before the path.", func(t *testing.T) {
			_, ast, err := runParser("include    some/path.journal\n")
			pruneMetadataFromAst(ast)

			assert.NoError(t, err)
			assert.Equal(t, &ledger.Journal{
				Entries: []ledger.Entry{
					&ledger.IncludeDirective{
						IncludePath: "some/path.journal",
					},
				},
			}, ast)
		})

		t.Run("Fail if an include does not have a path.", func(t *testing.T) {
			_, _, err := runParser("include\n")

			assert.Error(t, err)
		})
	})

	t.Run("Account directive", func(t *testing.T) {
		t.Run("Parses a file containing an account directive with multiple segments", func(t *testing.T) {
			_, ast, err := runParser("account assets:Cash:Checking\n")
			pruneMetadataFromAst(ast)

			assert.NoError(t, err)
			assert.Equal(t, &ledger.Journal{
				Entries: []ledger.Entry{
					&ledger.AccountDirective{
						AccountName: &ledger.AccountName{
							Segments: []string{"assets", "Cash", "Checking"},
						},
					},
				},
			}, ast)
		})

		t.Run("Parses a file containing an account directive with special characters and whitespace", func(t *testing.T) {
			_, ast, err := runParser("account assets:Cash:Che cking:Spe-ci_al\n")
			pruneMetadataFromAst(ast)

			assert.NoError(t, err)
			assert.Equal(t, &ledger.Journal{
				Entries: []ledger.Entry{
					&ledger.AccountDirective{
						AccountName: &ledger.AccountName{
							Segments: []string{"assets", "Cash", "Che cking", "Spe-ci_al"},
						},
					},
				},
			}, ast)
		})

		t.Run("Allows multiple spaces before the account name.", func(t *testing.T) {
			_, ast, err := runParser("account    assets:Cash\n")
			pruneMetadataFromAst(ast)

			assert.NoError(t, err)
			assert.Equal(t, &ledger.Journal{
				Entries: []ledger.Entry{
					&ledger.AccountDirective{
						AccountName: &ledger.AccountName{
							Segments: []string{"assets", "Cash"},
						},
					},
				},
			}, ast)
		})

		t.Run("Returns an error if the account name is missing.", func(t *testing.T) {
			_, _, err := runParser("account\n")

			assert.Error(t, err)
		})
	})

	t.Run("Mixed", func(t *testing.T) {
		t.Run("Parses a journal file containing many different directives, postings and comments", func(t *testing.T) {
			_, ast, err := runParser(
				`; This is a cool journal file
# It includes many things
; This next line is empty, but contains whitespace:
    

include someLong/Pathof/things.journal
include also/there-are-no/inlinecomments/after.includes  ; this is part of the path

account assets:Cash:Checking
    ; indented comment
account expenses:Gro ce:ries  ; hehe

payee Some Cool Person

commodity EUR
    format 1,000.00 €

2024-11-25 ! (code) Payee | transaction reason  ; inline transction comment
    expenses:Groceries      1,234.56 €
    ! assets:Cash:Checking   -1,234.56 €  ; inline posting comment

2024-11-25 Payee | transaction reason
    (virtual:posting)      300 €
    [balanced:virtual:posting]   = 15 €
`)
			pruneMetadataFromAst(ast)

			assert.NoError(t, err)
			assert.Equal(t,
				&ledger.Journal{
					Entries: []ledger.Entry{
						&ledger.IncludeDirective{
							IncludePath: "someLong/Pathof/things.journal",
						},
						&ledger.IncludeDirective{
							IncludePath: "also/there-are-no/inlinecomments/after.includes  ; this is part of the path",
						},
						&ledger.AccountDirective{
							AccountName: &ledger.AccountName{
								Segments: []string{"assets", "Cash", "Checking"},
							},
						},
						&ledger.AccountDirective{
							AccountName: &ledger.AccountName{
								Segments: []string{"expenses", "Gro ce", "ries"},
							},
						},
						&ledger.RealPosting{
							PostingStatus: "",
							AccountName: &ledger.AccountName{
								Segments: []string{"expenses", "Groceries"},
							},
							Amount: "1,234.56 €",
						},
						&ledger.RealPosting{
							PostingStatus: "!",
							AccountName: &ledger.AccountName{
								Segments: []string{"assets", "Cash", "Checking"},
							},
							Amount: "-1,234.56 €",
						},
						&ledger.VirtualPosting{
							PostingStatus: "",
							AccountName: &ledger.AccountName{
								Segments: []string{"virtual", "posting"},
							},
							Amount: "300 €",
						},
						&ledger.VirtualBalancedPosting{
							PostingStatus: "",
							AccountName: &ledger.AccountName{
								Segments: []string{"balanced", "virtual", "posting"},
							},
							Amount: "= 15 €",
						},
					},
				},
				ast,
			)
		})

		t.Run("fails to parse invalid inputs.", func(t *testing.T) {
			invalidInputs := []string{
				"account assets:Cash:Checking  heckmeck\n",
				"  (foo\n",
			}

			for _, invalidInput := range invalidInputs {
				_, _, err := runParser(invalidInput)

				assert.Error(t, err)
			}
		})
	})
}
