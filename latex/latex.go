package latex

import (
	"errors"
	"unicode"
	"strconv"
)

func TransformLatex(latex string) latexTransFormResult {
	original := latex + " "
	info := latexTransformationInfo{
		current_string: "",
		openBraces: 0,
		mode: notOpen,
		envMode: noMathEnv,
		prevToken: token{
			ttype: none,
			tokenInfo: "",
		},
		commands: commandHandling{
			customCommands: GetCustomCommands(),
			commandReplacements: GetCommandReplacements(),
			usedCustomCommands: make(map[string]bool),
		},
		knownMathEnvirons: GetKnownMathEnvirons(),
		braceReplacement: make([]braceClosingData, 0),
		environmentStack: make([]string, 0),
		environmentReplacements: GetEnvReplacements(),
		bracketReplacement: nil,
		html: false,
		logMap: make(map[string]int),
	}
	info.setHtmlIfNeeded(original)
	for _, char := range original {
		handleCharacter(char, &info)
	}
	prelude_string := info.preamble()
	logString := ""
	for key, val := range info.logMap {
		logString += strconv.Itoa(val) + "x " + string(key) + "\n"
	}
	infoStr := ""
	if info.html {
		infoStr = "Output contains HTML.\nInput in Moodle as source code (Ansicht -> Quellcode)!"
	}
	return latexTransFormResult{
		prelude_string + info.current_string[0:len(info.current_string) - 1],
		[]string{},
		true,
		"",
		logString,
		infoStr,
	}
}

func braceCheck(info *latexTransformationInfo, char rune) bool {
	if char == '{' || char == '}' {
		if char == '{' {
			info.incrBraces()
		} else if char == '}' {
			info.decrBraces()
		}
		if info.anyClosingBraceAction() {
			action := info.getClosingBraceAction()
			if action.depth == info.getOpenBraces() {
				if action.escape {
					info.addToOutputString(action.replacement)
				} else {
					info.addRawToOutputString(action.replacement)
				}
				info.popClosingBraceAction()
				return true
			}
		}
	} else if char == ']' {
		repl := info.extractBracketReplacement()
		if repl != nil {
			if repl.escape {
				info.addToOutputString(repl.replacement)
			} else {
				info.addRawToOutputString(repl.replacement)
			}
			return true
		}
	}
	return false
}

func checkForComment(info *latexTransformationInfo, char rune) {
	if char == '%' {
		info.log("Removed comment")
		info.setPrevToken(token{comment, ""})
	}
}

func handleCharacter(char rune, info *latexTransformationInfo) error {
	switch info.getTokenType() {
	case dollar:
		err := handlePrevDollarChar(char, info)
		if err != nil {
			return err
		}
	case none:
		err := handlePrevNoTokenChar(char, info)
		if err != nil {
			return err
		}
	case backslash:
		err := handlePrevBackslash(char, info)
		if err != nil {
			return err
		}
	case comment:
		if char == '\n' {
			info.setPrevToken(token{none, ""})
		}
	case backslashOngoing:
		err := handlePrevBackslashOngoing(char, info)
		if err != nil {
			return err
		}
	case environOpen:
		err := handlePrevEnvironOpen(char, info)
		if err != nil {
			return err
		}
	case environClose:
		err := handlePrevEnvironClose(char, info)
		if err != nil {
			return err
		}
	}
	return nil
}

func handlePrevDollarChar(char rune, info *latexTransformationInfo) error {
	info.setPrevToken(token{none, ""})
	switch char {
	case '$':
		switch info.getMathMode() {
		case notOpen:
			info.setMathMode(block)
			info.addToOutputString("\\[")
			info.log("Replaced $...$ with \\[...\\]")
		case inline:
			return errors.New("math error: $$ after open $")
		case block:
			info.setMathMode(notOpen)
			info.addToOutputString("\\]")
		}
	default:
		switch info.getMathMode() {
		case notOpen:
			info.setMathMode(inline)
			info.addToOutputString("\\(")
			info.log("Replaced $...$ with \\(...\\)")
		case inline:
			info.setMathMode(notOpen)
			info.addToOutputString("\\)")
		case block:
			return errors.New("math error: $ after open $$")
		}
		switch char {
		case '\\':
			info.setPrevToken(token{backslash, ""})
		default:
			checkForComment(info, char)
			handleCharacter(char, info)
		}
	}
	return nil
}

func handlePrevNoTokenChar(char rune, info *latexTransformationInfo) error {
	switch char {
	case '$':
		info.setPrevToken(token{dollar, ""})
	case '\\':
		info.setPrevToken(token{backslash, ""})
	case '%':
		checkForComment(info, char)
	default:
		action := braceCheck(info, char)
		if !action {
			info.addToOutputString(string(char))
		}
	}
	return nil
}

func handlePrevBackslash(char rune, info *latexTransformationInfo) error {
	if char == '\\' {
		if !info.isMathModeActive() && info.getEnv() == noMathEnv {
			info.log("Wrapped newline \\\\ in \\( \\)")
			info.addToOutputString("\\(\\\\ \\)")
		} else {
			info.addToOutputString("\\\\")
		}
		info.setPrevToken(token{none, ""})
	} else {
		switch char {
		case '$', '{', '}', '%':
			// no brace check or math mode entering, as character is escaped
			info.setPrevToken(token{none, ""})
			info.addToOutputString("\\" + string(char))
		case '(', '[':
			if !info.isMathModeActive() {
				if char == '(' {
					info.setMathMode(inline)
				} else {
					info.setMathMode(block)
				}
				info.setPrevToken(token{none, ""})
				info.addToOutputString("\\" + string(char))
			} else {
				return errors.New("unexpected math mode opening \\" + string(char) + " in math mode")
			}
		case ')', ']':
			if !info.isMathModeActive() {
				return errors.New("unexpected closure \\" + string(char) + " of math mode")
			} else {
				if char == ')' {
					if info.getMathMode() == block {
						return errors.New("mismatched closure of math mode: \\)")
					}
					info.setMathMode(notOpen)
					info.setPrevToken(token{none, ""})
					info.addToOutputString("\\" + string(char))
				} else {
					if info.getMathMode() == inline {
						return errors.New("mismatched closure of math mode: \\]")
					}
					info.setMathMode(notOpen)
					info.setPrevToken(token{none, ""})
					info.addToOutputString("\\" + string(char))
				}
			}
		default:
			if unicode.IsLetter(char) {
				info.setPrevToken(token{backslashOngoing, string(char)})
			} else {
				info.addToOutputString("\\" + string(char))
				info.setPrevToken(token{none, ""})
			}
		}
	}
	return nil
}

func handlePrevBackslashOngoing(char rune, info *latexTransformationInfo) error {
	if unicode.IsLetter(char) {
		info.addTokenInfo(string(char))
	} else {
		var repl commandReplacement = commandReplacement{}
		repl_valid := false
		command := string(info.getTokenInfo())
		oldCommand := command
		_, contains := info.customCommands()[command]
		if contains {
			info.addCommandUsage(command)
		} else {
			replac, ok := info.getCommandReplacement(command)
			repl = replac
			if ok {
				repl_valid = true
				if repl.argCommand {
					if char != '{' {
						return errors.New("expected { for command "+ string(command))
					}
				} else if repl.optArgCommand {
					if char != '[' {
						return errors.New("expected [ for command "+ string(command))
					}
				} else {
					command = repl.leftRepl[1:]
				}
			}
		}
		if command == "begin" {
			braceCheck(info, char)
			info.setPrevToken(token{environOpen, ""})
		} else if command == "end" {
			braceCheck(info, char)
			info.setPrevToken(token{environClose, ""})
		} else {
			if repl_valid {
				if repl.argCommand {
					if repl.rightRepl != "" {
						info.log("Replaced \\" + string(oldCommand) + "{...} with " + repl.leftRepl + "..." + repl.rightRepl)
					} else {
						info.log("Replaced \\" + string(oldCommand) + "{...} with " + repl.leftRepl)
					}
				} else if repl.optArgCommand {
					if repl.rightRepl != "" {
						info.log("Replaced \\" + string(oldCommand) + "[...] with " + repl.leftRepl +"..." + repl.rightRepl)
					} else {
						info.log("Replaced \\" + string(oldCommand) + "[...] with" + repl.leftRepl)
					}
				} else {
					info.log("Replaced \\" + string(oldCommand) + " with " + repl.leftRepl)
				}
				if repl.escapeRepl {
					info.addToOutputString(repl.leftRepl)
				} else {
					info.addRawToOutputString(repl.leftRepl)
				}
				info.setPrevToken(token{none, ""})
				if repl.argCommand {
					info.pushClosingBraceAction(braceClosingData{repl.escapeRepl, info.getOpenBraces(), repl.rightRepl})
					braceCheck(info, char)
				} else if repl.optArgCommand {
					info.setBracketReplacement(bracketClosingData{repl.escapeRepl, repl.rightRepl})
					braceCheck(info, char)
				} else {
					err := handleCharacter(char, info)
					if err != nil {
						return err
					}
				}
			} else {
				info.addToOutputString("\\" + command)
				info.setPrevToken(token{none, ""})
				err := handleCharacter(char, info)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func handlePrevEnvironOpen(char rune, info *latexTransformationInfo) error {
	braceCheck(info, char)
	if char == '}' {
		environ := info.getTokenInfo()
		info.addEnvironment(environ)
		repl, ok := info.getEnvRepl(environ)
		if ok {
			info.log("Replaced environment " + environ + " with " + string(repl.leftRepl) + "..." + string(repl.rightRepl))
			if repl.escapeRepl {
				info.addToOutputString(repl.leftRepl)
			} else {
				info.addRawToOutputString(repl.leftRepl)
			}
		}
		info.setPrevToken(token{none, ""})
		contains := info.getKnownMathEnvirons(environ)
		if contains && info.countMathEnvs() <= 1 && info.mode == notOpen {
			if !ok {
				info.log("Wrapped environment " + string(environ) + " in \\( \\)")
				info.addToOutputString("\\(\\begin{" + environ + "}")
			}
			info.setEnv(mathEnv)
		} else {
			if !ok {
				info.addToOutputString("\\begin{" + environ + "}")
			}
		}
	} else {
		info.addTokenInfo(string(char))
	}
	return nil
}

func handlePrevEnvironClose(char rune, info *latexTransformationInfo) error {
	braceCheck(info, char)
	if char == '}' {
		environ := info.getTokenInfo()
		err := info.popEnvironment(environ)
		if err != nil {
			return err
		}
		repl, ok := info.getEnvRepl(environ)
		if ok {
			if repl.escapeRepl {
				info.addToOutputString(repl.rightRepl)
			} else {
				info.addRawToOutputString(repl.rightRepl)
			}
		}
		info.setPrevToken(token{none, ""})
		contains := info.getKnownMathEnvirons(environ)
		if contains && info.countMathEnvs() == 0 && info.mode == notOpen {
			if !ok {
				info.addToOutputString("\\end{" + environ + "}\\)")
			}
			info.setEnv(noMathEnv)
		} else {
			if !ok {
				info.addToOutputString("\\end{" + environ + "}")
			}
		}
	} else {
		info.addTokenInfo(string(char))
	}
	return nil
}