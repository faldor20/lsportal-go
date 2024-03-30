package lsportal

// InclusionLanguage represents the configuration for a language that can be
// included in another file (e.g., Python code blocks in a Markdown file).
type InclusionLanguage struct {
	//The command to run the language server.
	Cmd string

	// Regex that captures regions where the inclusion language is.
	Regex string

	// Regex that detects parts of the inclusion language that
	// should be excluded from completion and other
	// language server features (e.g., string interpolation).
	ExclusionRegex string

	// The file extension associated with the inclusion language.
	Extension string

	// Arguments to be passed to the language server.
	LspArgs []string
}
