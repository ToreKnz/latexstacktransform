package latex

type envReplacement struct {
	escapeRepl 	bool
	leftRepl	string
	rightRepl	string
	innerRepl 	map[string]commandReplacement
}

func GetEnvReplacements() map[string]envReplacement {
	return map[string]envReplacement {
		"enumerate": {
			escapeRepl: false,
			leftRepl: "<ol>",
			rightRepl: "</ol>",
			innerRepl: map[string]commandReplacement{
				"item": {
					argCommand: false,
					optArgCommand: false,
					leftRepl: "<li>",
					rightRepl: "",
				},
			},
		},
		"itemize": {
			escapeRepl: false,
			leftRepl: "<ul>",
			rightRepl: "</ul>",
			innerRepl: map[string]commandReplacement{
				"item": {
					argCommand: false,
					optArgCommand: false,
					leftRepl: "<li>",
					rightRepl: "",
				},
			},
		}, 
		"description": {
			escapeRepl: false,
			leftRepl: "",
			rightRepl: "",
			innerRepl: map[string]commandReplacement{
				"item": {
					argCommand: false,
					optArgCommand: true,
					leftRepl: "",
					rightRepl: "",
				},
			},
		},
	}
}

func GetKnownMathEnvirons() map[string]bool {
	return map[string]bool {
		"align": true,
		"align*": true,
		"aligned": true,
		"equation": true,
		"equation*": true,
	}
}