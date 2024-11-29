package ledger

import "io/fs"

func ResolveIncludes(journal *Journal, parser *JournalParser, fs fs.FS) (*Journal, error) {
	newJournal := Journal{
		Entries: make([]Entry, 0),
	}

	for _, entry := range journal.Entries {
		switch entry := entry.(type) {
		case *IncludeDirective:
			includePath := entry.IncludePath
			file, err := fs.Open(includePath)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			includeJournal, err := parser.Parse(includePath, file)
			if err != nil {
				return nil, err
			}
			resolvedIncludeJournal, err := ResolveIncludes(includeJournal, parser, fs)
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
