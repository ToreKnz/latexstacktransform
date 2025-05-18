package latex

import (
	"errors"
	"strconv"
	"strings"
)

type latexTransFormResult struct {
	Transformed 	string
	OperationsLog 	[]string
	Success 		bool
	ErrorMessage 	string
	Log				string
	Info			string
}

type mathModeOpen int
const (
	notOpen	mathModeOpen = iota
	inline
	block
)

type environMode int
const (
	mathEnv environMode = iota
	noMathEnv
)

type braceClosingData struct {
	escape 		bool
	depth		int
	replacement	string
}

type bracketClosingData struct {
	escape 		bool
	replacement	string
}

type latexTransformationInfo struct {
	current_string			string
	openBraces				int
	mode 					mathModeOpen
	envMode 				environMode
	prevToken				token
	commands 				commandHandling
	knownMathEnvirons   	map[string]bool
	braceReplacement 		[]braceClosingData
	environmentStack 		[]string
	environmentReplacements map[string]envReplacement
	bracketReplacement		*bracketClosingData
	html					bool
	logMap						map[string]int
}

func (l *latexTransformationInfo) log(s string) {
	val, ok := l.logMap[s]
	if !ok {
		l.logMap[s] = 1
	} else {
		l.logMap[s] = val + 1
	}
}

func (l *latexTransformationInfo) getEnvRepl(env string) (envReplacement, bool) {
	val, ok := l.environmentReplacements[env]
	return val, ok
}

func (l *latexTransformationInfo) getEnvCommandRepl(env string, command string) (commandReplacement, bool) {
	repl, ok := l.environmentReplacements[env]
	if !ok {
		return commandReplacement{}, false
	}
	inner_repl, ok := repl.innerRepl[command]
	if !ok {
		return commandReplacement{}, false
	}
	return inner_repl, true
}

func (l *latexTransformationInfo) extractBracketReplacement() *bracketClosingData {
	if l.bracketReplacement != nil {
		res := l.bracketReplacement
		l.bracketReplacement = nil
		return res
	}
	return nil
}

func (l *latexTransformationInfo) setBracketReplacement(b bracketClosingData) {
	l.bracketReplacement = &b
}

func (l *latexTransformationInfo) setHtmlIfNeeded(original string) {
	l.html = strings.Contains(original, "\\begin{enumerate}") || strings.Contains(original, "\\begin{itemize}")
}

func (l *latexTransformationInfo) getEnv() environMode {
	return l.envMode
}

func (l *latexTransformationInfo) setEnv(env environMode) {
	l.envMode = env
}

func (l *latexTransformationInfo) addToOutputString(text string) {
	if l.html {
		escaped := strings.Replace(text, "&", "&amp;", -1)
		escaped = strings.Replace(escaped, "<", "&lt;", -1)
		escaped = strings.Replace(escaped, ">", "&gt;", -1)
		escaped = strings.Replace(escaped, "\n", "<br>\n", -1)
		l.current_string += escaped
	} else {
		l.current_string += text
	}
}

func (l *latexTransformationInfo) addRawToOutputString(text string) {
		l.current_string += text
}

func (l *latexTransformationInfo) anyClosingBraceAction() bool {
	return len(l.braceReplacement) > 0
}

func (l *latexTransformationInfo) getClosingBraceAction() braceClosingData {
	return l.braceReplacement[len(l.braceReplacement)-1]
}

func (l *latexTransformationInfo) pushClosingBraceAction(action braceClosingData) {
	l.braceReplacement = append(l.braceReplacement, action)
}

func (l *latexTransformationInfo) popClosingBraceAction() {
	l.braceReplacement = l.braceReplacement[0:len(l.braceReplacement) - 1]
}

func (l *latexTransformationInfo) incrBraces() {
	l.openBraces += 1
}

func (l *latexTransformationInfo) decrBraces() {
	l.openBraces -= 1
}

func (l *latexTransformationInfo) getOpenBraces() int {
	return l.openBraces
}

func (l *latexTransformationInfo) getKnownMathEnvirons(environ string) bool {
	_, ok := l.knownMathEnvirons[environ]
	return ok
}

func (l *latexTransformationInfo) countMathEnvs() int {
	count := 0
	for _, env := range l.environmentStack {
		if l.getKnownMathEnvirons(env) {
			count += 1
		}
	}
	return count
}

func (l *latexTransformationInfo) customCommands() map[string]string {
	return l.commands.customCommands
}

func (l *latexTransformationInfo) addCommandUsage(command string) {
	l.commands.usedCustomCommands[command] = true
}

func (l *latexTransformationInfo) getCommandReplacement(command string) (commandReplacement, bool) {
	val, ok := l.commands.commandReplacements[command]	
	if !ok {
		env, envExits := l.getCurrEnv()
		if !envExits {
			return commandReplacement{}, false
		}
		// special case enumerate for convenience
		if env == "enumerate" && command == "item" {
			enumerateCount := 0
			for _, environ := range l.environmentStack {
				if environ == "enumerate" {
					enumerateCount += 1
				}
			}
			if enumerateCount == 1 {
				return commandReplacement{false, false, false, "<li type=\"a\">)", ""}, true
			} else if enumerateCount == 2 {
				return commandReplacement{false, false, false, "<li type=\"i\">)", ""}, true
			} else {
				return commandReplacement{false, false, false, "<li type=\"A\">)", ""}, true
			}
		}
		repl, found := l.getEnvCommandRepl(env, command)
		return repl, found
	}
	return val, ok
}

func (l *latexTransformationInfo) addEnvironment(env string) {
	l.environmentStack = append(l.environmentStack, env)
}

func (l *latexTransformationInfo) getCurrEnv() (string, bool) {
	if len(l.environmentStack) == 0 {
		return "", false
	}
	return l.environmentStack[len(l.environmentStack)-1], true
}

func (l *latexTransformationInfo) popEnvironment(env string) error {
	if (len(l.environmentStack) == 0) {
		return errors.New("unexpected environment closure: " + string(env))
	}
	lastEnv := l.environmentStack[len(l.environmentStack)-1]
	if lastEnv != env {
		return errors.New("unexpected environment closure: " + string(env) +", last open environment: " + lastEnv)
	}
	l.environmentStack = l.environmentStack[0:len(l.environmentStack) - 1]
	return nil
}

func (l *latexTransformationInfo) setMathMode(mode mathModeOpen) {
	l.mode = mode
}

func (l *latexTransformationInfo) getMathMode() mathModeOpen {
	return l.mode
}

func (l *latexTransformationInfo) isMathModeActive() bool {
	return l.mode != notOpen
}

func (l *latexTransformationInfo) preamble() string {
	return CreateCustomCommandPreamble(&l.commands.usedCustomCommands, l)
}

func (l *latexTransformationInfo) setPrevToken(tk token) {
	if tk.ttype != backslashOngoing && tk.ttype != environClose && tk.ttype != environOpen && tk.tokenInfo != "" {
		panic("Token info should be empty! Token of type " + tk.ttype.toString() + " does not contain info!")
	}
	l.prevToken = tk
}

func (l *latexTransformationInfo) getTokenInfo() string {
	if l.prevToken.ttype != backslashOngoing && l.prevToken.ttype != environClose && l.prevToken.ttype != environOpen {
		panic("Token info shoud not be requested for token of type " + l.prevToken.ttype.toString() + "!")
	}
	return l.prevToken.tokenInfo
}

func (l *latexTransformationInfo) addTokenInfo(s string) {
	if l.prevToken.ttype != backslashOngoing && l.prevToken.ttype != environClose && l.prevToken.ttype != environOpen {
		panic("Token info shoud not be added for token of type " + l.prevToken.ttype.toString() + "!")
	}
	l.prevToken.tokenInfo += s
}

func (l *latexTransformationInfo) getTokenType() tokenType {
	return l.prevToken.ttype
}

type commandHandling struct {
	commandReplacements map[string]commandReplacement
	customCommands 		map[string]string
	usedCustomCommands 	map[string]bool
}

type tokenType int
const (
	backslash tokenType = iota
	backslashOngoing
	environOpen
	environClose
	dollar
	comment
	none
)

func (t tokenType) toString() string {
	switch t {
	case backslash:
		return "backslash"
	case backslashOngoing:
		return "backslashOngoing"
	case environOpen:
		return "environOpen"
	case environClose:
		return "environClose"
	case dollar:
		return "dollar"
	case comment:
		return "comment"
	case none:
		return "none"
	default:
		panic("unknwon token of index: " + strconv.Itoa(int(t)))
	}
}

type token struct {
	ttype			tokenType
	tokenInfo		string
}