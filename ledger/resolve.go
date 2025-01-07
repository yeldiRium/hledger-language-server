package ledger

import (
	"io/fs"
	"path"
)

func ResolveIncludes(journal *Journal, journalFilePath string, parser *JournalParser, fs fs.FS) (*Journal, error) {
	newJournal := Journal{
		Entries: make([]Entry, 0),
	}
	journalDir := path.Dir(journalFilePath)

	for _, entry := range journal.Entries {
		switch entry := entry.(type) {
		case *IncludeDirective:
			includePath := entry.IncludePath
			if !path.IsAbs(includePath) {
				includePath = path.Join(journalDir, includePath)
			}
			if path.IsAbs(includePath) {
				includePath = includePath[1:]
			}

			file, err := fs.Open(includePath)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			includeJournal, err := parser.Parse(includePath, file)
			if err != nil {
				return nil, err
			}
			resolvedIncludeJournal, err := ResolveIncludes(includeJournal, journalFilePath, parser, fs)
			if err != nil {
				return nil, err
			}

			for _, entry := range resolvedIncludeJournal.Entries {
				newJournal.Entries = append(newJournal.Entries, entry)
			}
		default:
			newJournal.Entries = append(newJournal.Entries, entry)
		}
	}

	return &newJournal, nil
}
