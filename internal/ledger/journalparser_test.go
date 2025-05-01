package ledger

import (
	"testing"
)

func TestJournalParser(t *testing.T) {
	t.Run("General format", func(t *testing.T) {
		t.Run("Parses a file containing only newlines.", func(t *testing.T) {
			AssertParser(
				t,
				NewJournalParser(),
				ParserInput("\n\n\n"),
				ExpectAst(&Journal{}),
			)
		})
	})

	t.Run("Include directive", func(t *testing.T) {
		t.Run("Parses a file containing an include directive with a path.", func(t *testing.T) {
			AssertParser(
				t,
				NewJournalParser(),
				ParserInput("include some/path.journal\n"),
				ExpectAst(&Journal{
					Entries: []Entry{
						&IncludeDirective{
							IncludePath: "some/path.journal",
						},
					},
				}),
			)
		})

		t.Run("Allows multiple spaces before the path.", func(t *testing.T) {
			AssertParser(
				t,
				NewJournalParser(),
				ParserInput("include    some/path.journal\n"),
				ExpectAst(&Journal{
					Entries: []Entry{
						&IncludeDirective{
							IncludePath: "some/path.journal",
						},
					},
				}),
			)
		})

		t.Run("Fail if an include does not have a path.", func(t *testing.T) {
			AssertParserFails(
				t,
				NewJournalParser(),
				"include\n",
			)
		})
	})

	t.Run("Account directive", func(t *testing.T) {
		t.Run("Parses a file containing an account directive with multiple segments", func(t *testing.T) {
			AssertParser(
				t,
				NewJournalParser(),
				ParserInput("account assets:Cash:Checking\n"),
				ExpectAst(&Journal{
					Entries: []Entry{
						&AccountDirective{
							AccountName: &AccountName{
								Segments: []string{"assets", "Cash", "Checking"},
							},
						},
					},
				}),
			)
		})

		t.Run("Parses a file containing an account directive with special characters and whitespace", func(t *testing.T) {
			AssertParser(
				t,
				NewJournalParser(),
				ParserInput("account assets:Cash:Che cking:Spe-ci_al\n"),
				ExpectAst(&Journal{
					Entries: []Entry{
						&AccountDirective{
							AccountName: &AccountName{
								Segments: []string{"assets", "Cash", "Che cking", "Spe-ci_al"},
							},
						},
					},
				}),
			)
		})

		t.Run("Allows multiple spaces before the account name.", func(t *testing.T) {
			AssertParser(
				t,
				NewJournalParser(),
				ParserInput("account    assets:Cash\n"),
				ExpectAst(&Journal{
					Entries: []Entry{
						&AccountDirective{
							AccountName: &AccountName{
								Segments: []string{"assets", "Cash"},
							},
						},
					},
				}),
			)
		})

		t.Run("Returns an error if the account name is missing.", func(t *testing.T) {
			AssertParserFails(
				t,
				NewJournalParser(),
				"account\n",
			)
		})
	})

	t.Run("Mixed", func(t *testing.T) {
		t.Run("Parses a journal file containing many different directives, postings and comments", func(t *testing.T) {
			AssertParser(
				t,
				NewJournalParser(),
				ParserInput(`; This is a cool journal file
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
`),
				ExpectAst(
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
				),
			)
		})

		t.Run("fails to parse invalid inputs.", func(t *testing.T) {
			invalidInputs := []string{
				"account assets:Cash:Checking  heckmeck\n",
				"  (foo\n",
			}

			for _, invalidInput := range invalidInputs {
				AssertParserFails(
					t,
					NewJournalParser(),
					invalidInput,
				)
			}
		})
	})
}
