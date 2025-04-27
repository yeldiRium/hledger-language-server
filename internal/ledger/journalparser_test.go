package ledger

import (
	"encoding/json"
	"fmt"
	"testing"

	participleLexer "github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
)

func TestJournalParser(t *testing.T) {
	filename := "testFile"

	runParser := func(testFileContent string) ([]participleLexer.Token, *Journal, error) {
		lexer := NewJournalLexer()
		lex, _ := lexer.LexString(filename, testFileContent)
		tokens := make([]participleLexer.Token, 0)
		token, err := lex.Next()
		for err == nil && token.Type != participleLexer.EOF {
			tokens = append(tokens, token)
			token, err = lex.Next()
		}
		fmt.Printf("Tokens: %#v\n", tokens)

		parser := NewJournalParser()
		ast, err := parser.ParseString(filename, testFileContent)

		jsonAst, _ := json.Marshal(ast)
		fmt.Printf("AST: %s\n", jsonAst)

		return tokens, ast, err
	}

	t.Run("General format", func(t *testing.T) {
		t.Run("Parses a file containing only newlines.", func(t *testing.T) {
			_, ast, err := runParser("\n\n\n")
			assert.NoError(t, err)
			assert.Equal(t, &Journal{}, ast)
		})
	})

	t.Run("Include directive", func(t *testing.T) {
		t.Run("Parses a file containing an include directive with a path.", func(t *testing.T) {
			_, ast, err := runParser("include some/path.journal\n")
			pruneMetadataFromAst(ast)

			assert.NoError(t, err)
			assert.Equal(t, &Journal{
				Entries: []Entry{
					&IncludeDirective{
						IncludePath: "some/path.journal",
					},
				},
			}, ast)
		})

		t.Run("Allows multiple spaces before the path.", func(t *testing.T) {
			_, ast, err := runParser("include    some/path.journal\n")
			pruneMetadataFromAst(ast)

			assert.NoError(t, err)
			assert.Equal(t, &Journal{
				Entries: []Entry{
					&IncludeDirective{
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
			assert.Equal(t, &Journal{
				Entries: []Entry{
					&AccountDirective{
						AccountName: &AccountName{
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
			assert.Equal(t, &Journal{
				Entries: []Entry{
					&AccountDirective{
						AccountName: &AccountName{
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
			assert.Equal(t, &Journal{
				Entries: []Entry{
					&AccountDirective{
						AccountName: &AccountName{
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

2024-12-01 Payee | posting with trailing whitespace
    expenses:Groceries           
`)
			pruneMetadataFromAst(ast)

			assert.NoError(t, err)
			assert.Equal(t,
				&Journal{
					Entries: []Entry{
						&IncludeDirective{
							IncludePath: "someLong/Pathof/things.journal",
						},
						&IncludeDirective{
							IncludePath: "also/there-are-no/inlinecomments/after.includes  ; this is part of the path",
						},
						&AccountDirective{
							AccountName: &AccountName{
								Segments: []string{"assets", "Cash", "Checking"},
							},
						},
						&AccountDirective{
							AccountName: &AccountName{
								Segments: []string{"expenses", "Gro ce", "ries"},
							},
						},
						&RealPosting{
							PostingStatus: "",
							AccountName: &AccountName{
								Segments: []string{"expenses", "Groceries"},
							},
							Amount: "1,234.56 €",
						},
						&RealPosting{
							PostingStatus: "!",
							AccountName: &AccountName{
								Segments: []string{"assets", "Cash", "Checking"},
							},
							Amount: "-1,234.56 €",
						},
						&VirtualPosting{
							PostingStatus: "",
							AccountName: &AccountName{
								Segments: []string{"virtual", "posting"},
							},
							Amount: "300 €",
						},
						&VirtualBalancedPosting{
							PostingStatus: "",
							AccountName: &AccountName{
								Segments: []string{"balanced", "virtual", "posting"},
							},
							Amount: "= 15 €",
						},
						&RealPosting{
							AccountName: &AccountName{
								Segments: []string{"expenses", "Groceries"},
							},
							Amount: "",
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

func TestAccountName(t *testing.T) {
	t.Run("Prefixes", func(t *testing.T) {
		t.Run("returns an empty list if the account name has no segments.", func(t *testing.T) {
			accountName := &AccountName{
				Segments: []string{},
			}

			prefixes := accountName.Prefixes()

			assert.Equal(t, []AccountName{}, prefixes)
		})

		t.Run("returns all segment prefixes of the account name.", func(t *testing.T) {
			accountName := &AccountName{
				Segments: []string{"assets", "Cash", "Checking", "Whatever", "another layer"},
			}

			prefixes := accountName.Prefixes()

			assert.Equal(
				t,
				[]AccountName{
					{Segments: []string{"assets"}},
					{Segments: []string{"assets", "Cash"}},
					{Segments: []string{"assets", "Cash", "Checking"}},
					{Segments: []string{"assets", "Cash", "Checking", "Whatever"}},
					{Segments: []string{"assets", "Cash", "Checking", "Whatever", "another layer"}},
				},
				prefixes,
			)
		})
	})
}
