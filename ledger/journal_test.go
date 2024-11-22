package ledger_test

import (
	"fmt"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/yeldiRium/hledger-language-server/ledger"
)

func TestJournalParser(t *testing.T) {
	testParserWithFileContent := func(t *testing.T, testFileContent string, expectedValue interface{}) {
		lexer2 := ledger.MakeJournalLexer()
		lex, err := lexer2.LexString("testFile", testFileContent)
		tokens := make([]lexer.Token, 0)
		token, err := lex.Next()
		for err == nil && token.Type != lexer.EOF {
			tokens = append(tokens, token)
			token, err = lex.Next()
		}
		fmt.Printf("Tokens: %#v\n", tokens)

		parser := ledger.MakeJournalParser()
		value, err := parser.ParseString("testFile", testFileContent)

		fmt.Printf("AST: %#v\n", value)

		assert.NoError(t, err)
		assert.Equal(t, expectedValue, value)
	}
	// testParserFails := func(t *testing.T, testFileContent string, errorMessage string) {
	// 	parser := MakeJournalParser()
	// 	_, err := parser.ParseString("testFile", testFileContent)

	// 	assert.EqualError(t, err, errorMessage)
	// }

	t.Run("General format", func(t *testing.T) {
		t.Run("Parses a file containing only newlines.", func(t *testing.T) {
			testParserWithFileContent(t, "\n\n\n", &ledger.Journal{})
		})

		//t.Run("Fails if a file does not end with a newline", func(t *testing.T) {
		//	testParserFails(t, "", "testFile:1:6: unexpected token \"<EOF>\" (expected <newline>)")
		//})
	})

	t.Run("Account directive", func(t *testing.T) {
		t.Run("Parses a file containing an account directive with multiple segments", func(t *testing.T) {
			testParserWithFileContent(t, "account assets:Cash:Checking\n", &ledger.Journal{
				Entries: []ledger.Entry{
					&ledger.AccountDirective{
						AccountName: "assets:Cash:Checking",
						// AccountName: &AccountName{
						// 	Segments: []string{"assets", "Cash", "Checking"},
						// },
					},
				},
			})
		})

		t.Run("Parses a file containing an account directive with special characters and whitespace", func(t *testing.T) {
			testParserWithFileContent(t, "account assets:Cash:Che cking:Spe-ci_al\n", &ledger.Journal{
				Entries: []ledger.Entry{
					&ledger.AccountDirective{
						AccountName: "assets:Cash:Che cking:Spe-ci_al",
						// AccountName: &AccountName{
						// 	Segments: []string{"assets", "Cash", "Che cking", "Spe-ci_al"},
						// },
					},
				},
			})
		})

		// t.Run("Parses a file containing an account directive followed by an inline comment", func(t *testing.T) {
		// 	testParserWithFileContent(t, "account assets:Cash:Checking  ; hehe\n", &Journal{
		// 		Entries: []Entry{
		// 			&AccountDirective{
		// 				AccountName: &AccountName{
		// 					Segments: []string{"assets", "Cash", "Checking"},
		// 				},
		// 				Comment: &InlineComment{
		// 					String: "hehe",
		// 				},
		// 			},
		// 		},
		// 	})
		// })

		// t.Run("Fails on more than one consecutive space within an account name", func(t *testing.T) {
		// 	testParserFails(
		// 		t,
		// 		"account assets:Cash:Che  cking\n",
		// 		"testFile:1:24: unexpected token \" \" (expected <newline>)",
		// 	)
		// })
	})

	// t.Run("Payee directive", func(t *testing.T) {
	// 	t.Run("Parse a file containing a payee directive", func(t *testing.T) {
	// 		testParserWithFileContent(
	// 			t,
	// 			"payee Some Cool Person\n",
	// 			&Journal{
	// 				Entries: []Entry{
	// 					&PayeeDirective{
	// 						PayeeName: "Some Cool Person",
	// 					},
	// 				},
	// 			},
	// 		)
	// 	})

	// 	t.Run("Parse a file containing a payee directive including special chars", func(t *testing.T) {
	// 		testParserWithFileContent(
	// 			t,
	// 			"payee So:me\n",
	// 			&Journal{
	// 				Entries: []Entry{
	// 					&PayeeDirective{
	// 						PayeeName: "So:me",
	// 					},
	// 				},
	// 			},
	// 		)
	// 	})
	// })

// 	t.Run("Comment", func(t *testing.T) {
// 		t.Run("Parses a file containing a ;-comment", func(t *testing.T) {
// 			testParserWithFileContent(t, "; This is a ;-comment\n", &Journal{
// 				Entries: []Entry{
// 					&Comment{
// 						String: "This is a ;-comment",
// 					},
// 				},
// 			})
// 		})
// 
// 		t.Run("Parses a file containing a #-comment", func(t *testing.T) {
// 			testParserWithFileContent(t, "# This is a #-comment\n", &Journal{
// 				Entries: []Entry{
// 					&Comment{
// 						String: "This is a #-comment",
// 					},
// 				},
// 			})
// 		})
// 
// 		t.Run("Parses a file with indented comments", func(t *testing.T) {
// 			testParserWithFileContent(
// 				t,
// 				`; Without indentation
//     ; with some indentation
//        # even more indentation
// `,
// 				&Journal{
// 					Entries: []Entry{
// 						&Comment{
// 							String: "Without indentation",
// 						},
// 						&Comment{
// 							String: "with some indentation",
// 						},
// 						&Comment{
// 							String: "even more indentation",
// 						},
// 					},
// 				},
// 			)
// 		})
// 	})

// 	t.Run("Transaction", func(t *testing.T) {
// 		t.Run("Parses a minimal transaction line.", func(t *testing.T) {
// 			testParserWithFileContent(
// 				t,
// 				`2020-01-01
// `,
// 				&Journal{
// 					Entries: []Entry{
// 						&Transaction{
// 							Date: "2020-01-01",
// 						},
// 					},
// 				},
// 			)
// 		})
// 
// 		t.Run("Parses the most common kind of transaction line.", func(t *testing.T) {
// 			testParserWithFileContent(
// 				t,
// 				`2020-01-01 Some Payee | transaction reason
// `,
// 				&Journal{
// 					Entries: []Entry{
// 						&Transaction{
// 							Date:        "2020-01-01",
// 							Payee:       "Some Payee",
// 							Description: "transaction reason",
// 						},
// 					},
// 				},
// 			)
// 		})
// 
// 		t.Run("Parses a full transaction line with all features.", func(t *testing.T) {
// 			testParserWithFileContent(
// 				t,
// 				`2020-01-01 ! (code) Payee | transaction reason  ; inline comment
// `,
// 				&Journal{
// 					Entries: []Entry{
// 						&Transaction{
// 							Date:        "2020-01-01",
// 							Status:      "!",
// 							Code:        "code",
// 							Payee:       "Payee",
// 							Description: "transaction reason",
// 							Comment: &InlineComment{
// 								String: "inline comment",
// 							},
// 						},
// 					},
// 				},
// 			)
// 		})
// 
// 		t.Run("Parses a pending transaction.", func(t *testing.T) {
// 			testParserWithFileContent(
// 				t,
// 				`2020-01-01 !
// `,
// 				&Journal{
// 					Entries: []Entry{
// 						&Transaction{
// 							Date:   "2020-01-01",
// 							Status: "!",
// 						},
// 					},
// 				},
// 			)
// 		})
// 
// 		t.Run("Parses a cleared transaction.", func(t *testing.T) {
// 			testParserWithFileContent(
// 				t,
// 				`2020-01-01 *
// `,
// 				&Journal{
// 					Entries: []Entry{
// 						&Transaction{
// 							Date:   "2020-01-01",
// 							Status: "*",
// 						},
// 					},
// 				},
// 			)
// 		})
// 	})
// 
// 	t.Run("Posting", func(t *testing.T) {
// 		t.Run("Parses a transaction with a minimal posting line.", func(t *testing.T) {
// 			testParserWithFileContent(
// 				t,
// 				`2024-11-15
//     assets:Cash:Checking
// `,
// 				&Journal{
// 					Entries: []Entry{
// 						&Transaction{
// 							Date: "2024-11-15",
// 							Postings: []*Posting{
// 								&Posting{
// 									AccountName: &AccountName{
// 										Segments: []string{"assets", "Cash", "Checking"},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			)
// 		})
// 
// 		t.Run("Parses a transaction with a full posting line.", func(t *testing.T) {
// 			testParserWithFileContent(
// 				t,
// 				`2024-11-15
//     assets:Cash:Checking    -1,234.56 €
// `,
// 				&Journal{
// 					Entries: []Entry{
// 						&Transaction{
// 							Date: "2024-11-15",
// 							Postings: []*Posting{
// 								&Posting{
// 									AccountName: &AccountName{
// 										Segments: []string{"assets", "Cash", "Checking"},
// 									},
// 									Amount: "-1,234.56 €",
// 								},
// 							},
// 						},
// 					},
// 				},
// 			)
// 		})
// 
// 		t.Run("Parses a transaction with a posting line with an inline comment.", func(t *testing.T) {
// 			testParserWithFileContent(
// 				t,
// 				`2024-11-15
//     assets:Cash:Checking   -1,234.56 €  ; inline comment
// `,
// 				&Journal{
// 					Entries: []Entry{
// 						&Transaction{
// 							Date: "2024-11-15",
// 						},
// 					},
// 				},
// 			)
// 		})
// 
// 		t.Run("Parses a transaction with multiple postings.", func(t *testing.T) {
// 			testParserWithFileContent(
// 				t,
// 				`2024-11-15
//     expenses:Groceries      1,234.56 €
//     assets:Cash:Checking   -1,234.56 €  ; inline comment
// `,
// 				&Journal{
// 					Entries: []Entry{
// 						&Transaction{
// 							Date: "2024-11-15",
// 						},
// 					},
// 				},
// 			)
// 		})
// 	})
// 
// 	t.Run("Mixed", func(t *testing.T) {
// 		t.Run("Parses a journal file containing many different directives, postings and comments", func(t *testing.T) {
// 			testParserWithFileContent(
// 				t,
// 				`; This is a cool journal file
// # It includes many things
// account assets:Cash:Checking
// account expenses:Gro ce:ries  ; hehe
// 
// payee Some Cool Person
// `,
// 				&Journal{
// 					Entries: []Entry{
// 						&Comment{
// 							String: "This is a cool journal file",
// 						},
// 						&Comment{
// 							String: "It includes many things",
// 						},
// 						&AccountDirective{
// 							AccountName: &AccountName{
// 								Segments: []string{"assets", "Cash", "Checking"},
// 							},
// 						},
// 						&AccountDirective{
// 							AccountName: &AccountName{
// 								Segments: []string{"expenses", "Gro ce", "ries"},
// 							},
// 							Comment: &InlineComment{
// 								String: "hehe",
// 							},
// 						},
// 						&PayeeDirective{
// 							PayeeName: "Some Cool Person",
// 						},
// 					},
// 				},
// 			)
// 		})
// 	})
}
