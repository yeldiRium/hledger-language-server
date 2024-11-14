package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJournalParser(t *testing.T) {
	testParserWithFileContent := func(t *testing.T, testFileContent string, expectedValue interface{}) {
		parser := MakeJournalParser()
		value, err := parser.ParseString("testFile", testFileContent)

		assert.NoError(t, err)
		assert.Equal(t, expectedValue, value)
	}
	testParserFails := func(t *testing.T, testFileContent string, errorMessage string) {
		parser := MakeJournalParser()
		_, err := parser.ParseString("testFile", testFileContent)

		assert.EqualError(t, err, errorMessage)
	}

	t.Run("General format", func(t *testing.T) {
		t.Run("Fails if a file does not end with a newline", func(t *testing.T) {
			testParserFails(t, "; foo", "testFile:1:6: unexpected token \"<EOF>\" (expected <newline>)")
		})
	})

	t.Run("Account directive", func(t *testing.T) {
		t.Run("Parses a file containing an account directive with multiple segments", func(t *testing.T) {
			testParserWithFileContent(t, "account assets:Cash:Checking\n", &Journal{
				Entries: []Entry{
					&AccountDirective{
						AccountName: "assets:Cash:Checking",
					},
				},
			})
		})

		t.Run("Parses a file containing an account directive with special characters and whitespace", func(t *testing.T) {
			testParserWithFileContent(t, "account assets:Cash:Che cking:Spe-ci_al\n", &Journal{
				Entries: []Entry{
					&AccountDirective{
						AccountName: "assets:Cash:Che cking:Spe-ci_al",
					},
				},
			})
		})

		t.Run("Parses a file containing an account directive followed by an inline comment", func(t *testing.T) {
			testParserWithFileContent(t, "account assets:Cash:Checking  ; hehe\n", &Journal{
				Entries: []Entry{
					&AccountDirective{
						AccountName: "assets:Cash:Checking",
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
				"testFile:1:25: unexpected token \" \" (expected <word>)",
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
							PayeeName: "Some Cool Person",
						},
					},
				},
			)
		})

		t.Run("Parse a file containing a payee directive including special chars", func(t *testing.T) {
			testParserWithFileContent(
				t,
				"payee So:me\n",
				&Journal{
					Entries: []Entry{
						&PayeeDirective{
							PayeeName: "So:me",
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

		t.Run("Parses a file with indented comments", func(t *testing.T) {
			testParserWithFileContent(
				t,
				`; Without indentation
    ; with some indentation
       # even more indentation
`,
				&Journal{
					Entries: []Entry{
						&Comment{
							String: "Without indentation",
						},
						&Comment{
							String: "with some indentation",
						},
						&Comment{
							String: "even more indentation",
						},
					},
				},
			)
		})
	})

	//	t.Run("Transaction", func(t *testing.T) {
	//		t.Run("Parses a simple transaction with two postings including all values.", func(t *testing.T) {
	//			testParserWithFileContent(
	//				t,
	//				`2020-01-01 Payee | transaction reason
	//    expenses:Groceries  15.23 €
	//    assets:Cash:Checking  -15.23 €
	//`,
	//				&Journal{},
	//			)
	//		})
	//	})

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
							AccountName: "assets:Cash:Checking",
						},
						&AccountDirective{
							AccountName: "expenses:Gro ce:ries",
							Comment: &InlineComment{
								String: "hehe",
							},
						},
						&PayeeDirective{
							PayeeName: "Some Cool Person",
						},
					},
				},
			)
		})
	})
}
