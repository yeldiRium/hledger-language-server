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

	t.Run("Parses a file containing an account directive with multiple segments", func(t *testing.T) {
		testParserWithFileContent(t, `account assets:Cash:Checking`, &AccountDirective{
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
		})
	})

	t.Run("Parses a file containing an account directive with special characters and whitespace", func(t *testing.T) {
		testParserWithFileContent(t, `account assets:Cash:Che cking:Spe-ci_al`, &AccountDirective{
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
		})
	})

	t.Run("Fails on more than one consecutive space within an account name", func(t *testing.T) {
		testParserFails(
			t,
			`account assets:Cash:Che  cking`,
			"testFile:1:24: unexpected token \"  \"",
		)
	})
}
