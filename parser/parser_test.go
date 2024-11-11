package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	testParserWithFileContent := func(t *testing.T, testFileContent string, expectedValue interface{}) {
		parser := MakeParser()
		value, err := parser.ParseString("testFile", testFileContent)

		assert.NoError(t, err)
		assert.Equal(t, expectedValue, value)
	}
	testParserFails := func(t *testing.T, testFileContent string, errorMessage string) {
		parser := MakeParser()
		_, err := parser.ParseString("testFile", testFileContent)

		assert.EqualError(t, err, errorMessage)
	}

	t.Run("Account directive", func(t *testing.T) {
		t.Run("Parses a file containing an account directive with multiple segments", func(t *testing.T) {
			testParserWithFileContent(t, "account assets:Cash:Checking\n", &Journal{
				Entries: []Entry{
					&AccountDirective{
						AccountName: &AccountName{
							Segments: []*AccountNameSegment{
								{
									String: "assets",
								},
								{
									String: "Cash",
								},
								{
									String: "Checking",
								},
							},
						},
					},
				},
			})
		})

		t.Run("Parses a file containing an account directive with special characters and whitespace", func(t *testing.T) {
			testParserWithFileContent(t, "account assets:Cash:Che cking:Spe-ci_al\n", &Journal{
				Entries: []Entry{
					&AccountDirective{
						AccountName: &AccountName{
							Segments: []*AccountNameSegment{
								{
									String: "assets",
								},
								{
									String: "Cash",
								},
								{
									String: "Che cking",
								},
								{
									String: "Spe-ci_al",
								},
							},
						},
					},
				},
			})
		})

		t.Run("Parses a file containing an account directive followed by an inline comment", func(t *testing.T) {
			testParserWithFileContent(t, "account assets:Cash:Checking  ; hehe\n", &Journal{
				Entries: []Entry{
					&AccountDirective{
						AccountName: &AccountName{
							Segments: []*AccountNameSegment{
								{
									String: "assets",
								},
								{
									String: "Cash",
								},
								{
									String: "Checking",
								},
							},
						},
						Comment: &InlineComment{
							String: "hehe",
						},
					},
				},
			})
		})

		t.Run("Fails on more than one consecutive space within an account name", func(t *testing.T) {
			testParserFails(
				t,
				"account assets:Cash:Che  cking\n",
				"testFile:1:24: unexpected token \"  \" (expected <eol>)",
			)
		})
	})

	t.Run("Payee directive", func(t *testing.T) {
		t.Run("Parse a file containing a payee directive", func(t *testing.T) {
			testParserWithFileContent(
				t,
				"payee Some Cool Person\n",
				&Journal{
					Entries: []Entry{
						&PayeeDirective{
							PayeeName: &PayeeName{
								String: "Some Cool Person",
							},
						},
					},
				},
			)
		})
	})

	t.Run("Comment", func(t *testing.T) {
		t.Run("Parses a file containing a ;-comment", func(t *testing.T) {
			testParserWithFileContent(t, "; This is a ;-comment\n", &Journal{
				Entries: []Entry{
					&Comment{
						String: "This is a ;-comment",
					},
				},
			})
		})

		t.Run("Parses a file containing a #-comment", func(t *testing.T) {
			testParserWithFileContent(t, "# This is a #-comment\n", &Journal{
				Entries: []Entry{
					&Comment{
						String: "This is a #-comment",
					},
				},
			})
		})
	})

	t.Run("Mixed", func(t *testing.T) {
		t.Run("Parses a journal file containing many different directives, postings and comments", func(t *testing.T) {
			testParserWithFileContent(
				t,
				`; This is a cool journal file
# It includes many things
account assets:Cash:Checking
account expenses:Gro ce:ries  ; hehe

payee Some Cool Person
`,
				&Journal{
					Entries: []Entry{
						&Comment{
							String: "This is a cool journal file",
						},
						&Comment{
							String: "It includes many things",
						},
						&AccountDirective{
							AccountName: &AccountName{
								Segments: []*AccountNameSegment{
									{String: "assets"},
									{String: "Cash"},
									{String: "Checking"},
								},
							},
						},
						&AccountDirective{
							AccountName: &AccountName{
								Segments: []*AccountNameSegment{
									{String: "expenses"},
									{String: "Gro ce"},
									{String: "ries"},
								},
							},
							Comment: &InlineComment{
								String: "hehe",
							},
						},
						&PayeeDirective{
							PayeeName: &PayeeName{
								String: "Some Cool Person",
							},
						},
					},
				},
			)
		})
	})
}
