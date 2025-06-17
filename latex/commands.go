package latex

func GetCustomCommands() map[string]string {
	return map[string]string{
		"abs": "\\newcommand{\\abs}[1]{\\left|#1\\right|}",
		"norm": "\\newcommand{\\norm}[1]{\\left|\\!\\left|#1\\right|\\!\\right|}",
		"energynorm": "\\newcommand{\\energynorm}[1]{\\left|\\!\\left|\\!\\left|#1\\right|\\!\\right|\\!\\right|}",
		"normone": "\\newcommand{\\normone}[1]{\\norm{#1}_1}",
		"normtwo": "\\newcommand{\\normtwo}[1]{\\norm{#1}_2}",
		"norminf": "\\newcommand{\\norminf}[1]{\\norm{#1}_\\infty}",
	}
}

func customCommandsInOrder() []string {
	return []string{"abs", "norm", "energynorm", "normone", "normtwo", "norminf"}
}

func customCommandDependencies() map[string][]string {
	return map[string][]string {
		"energynorm": {"norm"},
		"normone": {"norm"},
		"normtwo": {"norm"},
		"norminf": {"norm"},
	}
}

type commandReplacement struct {
	escapeRepl		bool
	argCommand		bool
	optArgCommand 	bool
	leftRepl		string
	rightRepl		string
}

func GetCommandReplacements() map[string]commandReplacement {
	return map[string]commandReplacement{
		"R": {true, false, false, "\\mathbb{R}", ""},
		"mbox": {true, true, false, "{", "}"},
		"Tilde": {true, true, false, "\\tilde{", "}"},
		"intertext": {true, true, false, "\\text{", "}\\\\ "},
	}
}

func CreateCustomCommandPreamble(usedCustomCommands *map[string]bool, info *latexTransformationInfo) string {
	dependencyMap := customCommandDependencies()
	for comm := range *usedCustomCommands {
		dependencies, ok := dependencyMap[comm]
		if ok {
			for _, dep := range dependencies {
				(*usedCustomCommands)[dep] = true
			}
		}
	}
	ordered_comms := customCommandsInOrder()
	command_map := GetCustomCommands()
	prelude_defs := make([]string, 0)
	for _, comm := range ordered_comms {
		_, ok := (*usedCustomCommands)[comm]
		if ok {
			def := command_map[comm]
			prelude_defs = append(prelude_defs, def)
			info.log("Included definition for " + string(comm))
		}
	}
	if len(prelude_defs) == 0 {
		return ""
	} else {
		prelude_string := "\\("
		for _, def := range prelude_defs {
			prelude_string += def + " "
		}
		prelude_string += "\\)"
		return prelude_string
	}
}