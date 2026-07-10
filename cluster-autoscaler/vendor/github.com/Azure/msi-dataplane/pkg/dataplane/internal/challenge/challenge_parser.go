// Code generated from Challenge.g4 by ANTLR 4.13.2. DO NOT EDIT.

package challenge // Challenge
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type ChallengeParser struct {
	*antlr.BaseParser
}

var ChallengeParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func challengeParserInit() {
	staticData := &ChallengeParserStaticData
	staticData.LiteralNames = []string{
		"", "'\\t'", "' '", "'!'", "'\"'", "'#'", "'$'", "'%'", "'&'", "'''",
		"'('", "')'", "'*'", "'+'", "','", "'-'", "'.'", "'/'", "", "':'", "';'",
		"'<'", "'='", "'>'", "'?'", "'@'", "", "'['", "'\\'", "']'", "'^'",
		"'_'", "'`'", "'{'", "'|'", "'}'", "'~'",
	}
	staticData.SymbolicNames = []string{
		"", "HTAB", "SP", "EXCLAMATION_MARK", "DQUOTE", "HASH", "DOLLAR", "PERCENT",
		"AMPERSAND", "SQUOTE", "OPEN_PARENS", "CLOSE_PARENS", "ASTERISK", "PLUS",
		"COMMA", "MINUS", "PERIOD", "SLASH", "DIGIT", "COLON", "SEMICOLON",
		"LESS_THAN", "EQUALS", "GREATER_THAN", "QUESTION", "AT", "ALPHA", "OPEN_BRACKET",
		"BACKSLASH", "CLOSE_BRACKET", "CARET", "UNDERSCORE", "GRAVE", "OPEN_BRACE",
		"PIPE", "CLOSE_BRACE", "TILDE", "EXTENDED_ASCII",
	}
	staticData.RuleNames = []string{
		"header", "challenge", "auth_scheme", "auth_params", "token68", "auth_param",
		"auth_lhs", "auth_rhs", "rws", "quoted_string", "qd_text", "quoted_pair",
		"token", "tchar", "vchar", "obs_text",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 37, 190, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		1, 0, 1, 0, 5, 0, 35, 8, 0, 10, 0, 12, 0, 38, 9, 0, 1, 0, 1, 0, 5, 0, 42,
		8, 0, 10, 0, 12, 0, 45, 9, 0, 1, 0, 5, 0, 48, 8, 0, 10, 0, 12, 0, 51, 9,
		0, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 57, 8, 1, 3, 1, 59, 8, 1, 5, 1, 61, 8,
		1, 10, 1, 12, 1, 64, 9, 1, 1, 2, 1, 2, 1, 3, 1, 3, 5, 3, 70, 8, 3, 10,
		3, 12, 3, 73, 9, 3, 1, 3, 1, 3, 5, 3, 77, 8, 3, 10, 3, 12, 3, 80, 9, 3,
		1, 3, 5, 3, 83, 8, 3, 10, 3, 12, 3, 86, 9, 3, 1, 4, 4, 4, 89, 8, 4, 11,
		4, 12, 4, 90, 1, 4, 5, 4, 94, 8, 4, 10, 4, 12, 4, 97, 9, 4, 1, 5, 1, 5,
		5, 5, 101, 8, 5, 10, 5, 12, 5, 104, 9, 5, 1, 5, 1, 5, 5, 5, 108, 8, 5,
		10, 5, 12, 5, 111, 9, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 7, 1, 7, 3, 7, 119,
		8, 7, 1, 8, 4, 8, 122, 8, 8, 11, 8, 12, 8, 123, 1, 9, 1, 9, 1, 9, 4, 9,
		129, 8, 9, 11, 9, 12, 9, 130, 1, 9, 1, 9, 1, 10, 1, 10, 1, 10, 1, 10, 1,
		10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10,
		1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1,
		10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 3, 10,
		170, 8, 10, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 3, 11, 177, 8, 11, 1, 12,
		4, 12, 180, 8, 12, 11, 12, 12, 12, 181, 1, 13, 1, 13, 1, 14, 1, 14, 1,
		15, 1, 15, 1, 15, 0, 0, 16, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22,
		24, 26, 28, 30, 0, 4, 1, 0, 1, 2, 5, 0, 13, 15, 17, 18, 26, 26, 31, 31,
		36, 36, 9, 0, 3, 3, 5, 9, 12, 13, 15, 16, 18, 18, 26, 26, 30, 32, 34, 34,
		36, 36, 1, 0, 3, 37, 228, 0, 32, 1, 0, 0, 0, 2, 52, 1, 0, 0, 0, 4, 65,
		1, 0, 0, 0, 6, 67, 1, 0, 0, 0, 8, 88, 1, 0, 0, 0, 10, 98, 1, 0, 0, 0, 12,
		114, 1, 0, 0, 0, 14, 118, 1, 0, 0, 0, 16, 121, 1, 0, 0, 0, 18, 125, 1,
		0, 0, 0, 20, 169, 1, 0, 0, 0, 22, 171, 1, 0, 0, 0, 24, 179, 1, 0, 0, 0,
		26, 183, 1, 0, 0, 0, 28, 185, 1, 0, 0, 0, 30, 187, 1, 0, 0, 0, 32, 49,
		3, 2, 1, 0, 33, 35, 7, 0, 0, 0, 34, 33, 1, 0, 0, 0, 35, 38, 1, 0, 0, 0,
		36, 34, 1, 0, 0, 0, 36, 37, 1, 0, 0, 0, 37, 39, 1, 0, 0, 0, 38, 36, 1,
		0, 0, 0, 39, 43, 5, 14, 0, 0, 40, 42, 7, 0, 0, 0, 41, 40, 1, 0, 0, 0, 42,
		45, 1, 0, 0, 0, 43, 41, 1, 0, 0, 0, 43, 44, 1, 0, 0, 0, 44, 46, 1, 0, 0,
		0, 45, 43, 1, 0, 0, 0, 46, 48, 3, 2, 1, 0, 47, 36, 1, 0, 0, 0, 48, 51,
		1, 0, 0, 0, 49, 47, 1, 0, 0, 0, 49, 50, 1, 0, 0, 0, 50, 1, 1, 0, 0, 0,
		51, 49, 1, 0, 0, 0, 52, 62, 3, 4, 2, 0, 53, 58, 5, 2, 0, 0, 54, 59, 3,
		8, 4, 0, 55, 57, 3, 6, 3, 0, 56, 55, 1, 0, 0, 0, 56, 57, 1, 0, 0, 0, 57,
		59, 1, 0, 0, 0, 58, 54, 1, 0, 0, 0, 58, 56, 1, 0, 0, 0, 59, 61, 1, 0, 0,
		0, 60, 53, 1, 0, 0, 0, 61, 64, 1, 0, 0, 0, 62, 60, 1, 0, 0, 0, 62, 63,
		1, 0, 0, 0, 63, 3, 1, 0, 0, 0, 64, 62, 1, 0, 0, 0, 65, 66, 3, 24, 12, 0,
		66, 5, 1, 0, 0, 0, 67, 84, 3, 10, 5, 0, 68, 70, 7, 0, 0, 0, 69, 68, 1,
		0, 0, 0, 70, 73, 1, 0, 0, 0, 71, 69, 1, 0, 0, 0, 71, 72, 1, 0, 0, 0, 72,
		74, 1, 0, 0, 0, 73, 71, 1, 0, 0, 0, 74, 78, 5, 14, 0, 0, 75, 77, 7, 0,
		0, 0, 76, 75, 1, 0, 0, 0, 77, 80, 1, 0, 0, 0, 78, 76, 1, 0, 0, 0, 78, 79,
		1, 0, 0, 0, 79, 81, 1, 0, 0, 0, 80, 78, 1, 0, 0, 0, 81, 83, 3, 10, 5, 0,
		82, 71, 1, 0, 0, 0, 83, 86, 1, 0, 0, 0, 84, 82, 1, 0, 0, 0, 84, 85, 1,
		0, 0, 0, 85, 7, 1, 0, 0, 0, 86, 84, 1, 0, 0, 0, 87, 89, 7, 1, 0, 0, 88,
		87, 1, 0, 0, 0, 89, 90, 1, 0, 0, 0, 90, 88, 1, 0, 0, 0, 90, 91, 1, 0, 0,
		0, 91, 95, 1, 0, 0, 0, 92, 94, 5, 22, 0, 0, 93, 92, 1, 0, 0, 0, 94, 97,
		1, 0, 0, 0, 95, 93, 1, 0, 0, 0, 95, 96, 1, 0, 0, 0, 96, 9, 1, 0, 0, 0,
		97, 95, 1, 0, 0, 0, 98, 102, 3, 12, 6, 0, 99, 101, 7, 0, 0, 0, 100, 99,
		1, 0, 0, 0, 101, 104, 1, 0, 0, 0, 102, 100, 1, 0, 0, 0, 102, 103, 1, 0,
		0, 0, 103, 105, 1, 0, 0, 0, 104, 102, 1, 0, 0, 0, 105, 109, 5, 22, 0, 0,
		106, 108, 7, 0, 0, 0, 107, 106, 1, 0, 0, 0, 108, 111, 1, 0, 0, 0, 109,
		107, 1, 0, 0, 0, 109, 110, 1, 0, 0, 0, 110, 112, 1, 0, 0, 0, 111, 109,
		1, 0, 0, 0, 112, 113, 3, 14, 7, 0, 113, 11, 1, 0, 0, 0, 114, 115, 3, 24,
		12, 0, 115, 13, 1, 0, 0, 0, 116, 119, 3, 24, 12, 0, 117, 119, 3, 18, 9,
		0, 118, 116, 1, 0, 0, 0, 118, 117, 1, 0, 0, 0, 119, 15, 1, 0, 0, 0, 120,
		122, 7, 0, 0, 0, 121, 120, 1, 0, 0, 0, 122, 123, 1, 0, 0, 0, 123, 121,
		1, 0, 0, 0, 123, 124, 1, 0, 0, 0, 124, 17, 1, 0, 0, 0, 125, 128, 5, 4,
		0, 0, 126, 129, 3, 20, 10, 0, 127, 129, 3, 22, 11, 0, 128, 126, 1, 0, 0,
		0, 128, 127, 1, 0, 0, 0, 129, 130, 1, 0, 0, 0, 130, 128, 1, 0, 0, 0, 130,
		131, 1, 0, 0, 0, 131, 132, 1, 0, 0, 0, 132, 133, 5, 4, 0, 0, 133, 19, 1,
		0, 0, 0, 134, 170, 5, 1, 0, 0, 135, 170, 5, 2, 0, 0, 136, 170, 5, 3, 0,
		0, 137, 170, 5, 5, 0, 0, 138, 170, 5, 6, 0, 0, 139, 170, 5, 7, 0, 0, 140,
		170, 5, 8, 0, 0, 141, 170, 5, 9, 0, 0, 142, 170, 5, 10, 0, 0, 143, 170,
		5, 11, 0, 0, 144, 170, 5, 12, 0, 0, 145, 170, 5, 13, 0, 0, 146, 170, 5,
		14, 0, 0, 147, 170, 5, 15, 0, 0, 148, 170, 5, 16, 0, 0, 149, 170, 5, 17,
		0, 0, 150, 170, 5, 18, 0, 0, 151, 170, 5, 19, 0, 0, 152, 170, 5, 20, 0,
		0, 153, 170, 5, 21, 0, 0, 154, 170, 5, 22, 0, 0, 155, 170, 5, 23, 0, 0,
		156, 170, 5, 24, 0, 0, 157, 170, 5, 25, 0, 0, 158, 170, 5, 27, 0, 0, 159,
		170, 5, 29, 0, 0, 160, 170, 5, 30, 0, 0, 161, 170, 5, 31, 0, 0, 162, 170,
		5, 32, 0, 0, 163, 170, 5, 26, 0, 0, 164, 170, 5, 33, 0, 0, 165, 170, 5,
		34, 0, 0, 166, 170, 5, 35, 0, 0, 167, 170, 5, 36, 0, 0, 168, 170, 3, 30,
		15, 0, 169, 134, 1, 0, 0, 0, 169, 135, 1, 0, 0, 0, 169, 136, 1, 0, 0, 0,
		169, 137, 1, 0, 0, 0, 169, 138, 1, 0, 0, 0, 169, 139, 1, 0, 0, 0, 169,
		140, 1, 0, 0, 0, 169, 141, 1, 0, 0, 0, 169, 142, 1, 0, 0, 0, 169, 143,
		1, 0, 0, 0, 169, 144, 1, 0, 0, 0, 169, 145, 1, 0, 0, 0, 169, 146, 1, 0,
		0, 0, 169, 147, 1, 0, 0, 0, 169, 148, 1, 0, 0, 0, 169, 149, 1, 0, 0, 0,
		169, 150, 1, 0, 0, 0, 169, 151, 1, 0, 0, 0, 169, 152, 1, 0, 0, 0, 169,
		153, 1, 0, 0, 0, 169, 154, 1, 0, 0, 0, 169, 155, 1, 0, 0, 0, 169, 156,
		1, 0, 0, 0, 169, 157, 1, 0, 0, 0, 169, 158, 1, 0, 0, 0, 169, 159, 1, 0,
		0, 0, 169, 160, 1, 0, 0, 0, 169, 161, 1, 0, 0, 0, 169, 162, 1, 0, 0, 0,
		169, 163, 1, 0, 0, 0, 169, 164, 1, 0, 0, 0, 169, 165, 1, 0, 0, 0, 169,
		166, 1, 0, 0, 0, 169, 167, 1, 0, 0, 0, 169, 168, 1, 0, 0, 0, 170, 21, 1,
		0, 0, 0, 171, 176, 5, 28, 0, 0, 172, 177, 5, 1, 0, 0, 173, 177, 5, 2, 0,
		0, 174, 177, 3, 28, 14, 0, 175, 177, 3, 30, 15, 0, 176, 172, 1, 0, 0, 0,
		176, 173, 1, 0, 0, 0, 176, 174, 1, 0, 0, 0, 176, 175, 1, 0, 0, 0, 177,
		23, 1, 0, 0, 0, 178, 180, 3, 26, 13, 0, 179, 178, 1, 0, 0, 0, 180, 181,
		1, 0, 0, 0, 181, 179, 1, 0, 0, 0, 181, 182, 1, 0, 0, 0, 182, 25, 1, 0,
		0, 0, 183, 184, 7, 2, 0, 0, 184, 27, 1, 0, 0, 0, 185, 186, 7, 3, 0, 0,
		186, 29, 1, 0, 0, 0, 187, 188, 5, 37, 0, 0, 188, 31, 1, 0, 0, 0, 20, 36,
		43, 49, 56, 58, 62, 71, 78, 84, 90, 95, 102, 109, 118, 123, 128, 130, 169,
		176, 181,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// ChallengeParserInit initializes any static state used to implement ChallengeParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewChallengeParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func ChallengeParserInit() {
	staticData := &ChallengeParserStaticData
	staticData.once.Do(challengeParserInit)
}

// NewChallengeParser produces a new parser instance for the optional input antlr.TokenStream.
func NewChallengeParser(input antlr.TokenStream) *ChallengeParser {
	ChallengeParserInit()
	this := new(ChallengeParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &ChallengeParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "Challenge.g4"

	return this
}

// ChallengeParser tokens.
const (
	ChallengeParserEOF              = antlr.TokenEOF
	ChallengeParserHTAB             = 1
	ChallengeParserSP               = 2
	ChallengeParserEXCLAMATION_MARK = 3
	ChallengeParserDQUOTE           = 4
	ChallengeParserHASH             = 5
	ChallengeParserDOLLAR           = 6
	ChallengeParserPERCENT          = 7
	ChallengeParserAMPERSAND        = 8
	ChallengeParserSQUOTE           = 9
	ChallengeParserOPEN_PARENS      = 10
	ChallengeParserCLOSE_PARENS     = 11
	ChallengeParserASTERISK         = 12
	ChallengeParserPLUS             = 13
	ChallengeParserCOMMA            = 14
	ChallengeParserMINUS            = 15
	ChallengeParserPERIOD           = 16
	ChallengeParserSLASH            = 17
	ChallengeParserDIGIT            = 18
	ChallengeParserCOLON            = 19
	ChallengeParserSEMICOLON        = 20
	ChallengeParserLESS_THAN        = 21
	ChallengeParserEQUALS           = 22
	ChallengeParserGREATER_THAN     = 23
	ChallengeParserQUESTION         = 24
	ChallengeParserAT               = 25
	ChallengeParserALPHA            = 26
	ChallengeParserOPEN_BRACKET     = 27
	ChallengeParserBACKSLASH        = 28
	ChallengeParserCLOSE_BRACKET    = 29
	ChallengeParserCARET            = 30
	ChallengeParserUNDERSCORE       = 31
	ChallengeParserGRAVE            = 32
	ChallengeParserOPEN_BRACE       = 33
	ChallengeParserPIPE             = 34
	ChallengeParserCLOSE_BRACE      = 35
	ChallengeParserTILDE            = 36
	ChallengeParserEXTENDED_ASCII   = 37
)

// ChallengeParser rules.
const (
	ChallengeParserRULE_header        = 0
	ChallengeParserRULE_challenge     = 1
	ChallengeParserRULE_auth_scheme   = 2
	ChallengeParserRULE_auth_params   = 3
	ChallengeParserRULE_token68       = 4
	ChallengeParserRULE_auth_param    = 5
	ChallengeParserRULE_auth_lhs      = 6
	ChallengeParserRULE_auth_rhs      = 7
	ChallengeParserRULE_rws           = 8
	ChallengeParserRULE_quoted_string = 9
	ChallengeParserRULE_qd_text       = 10
	ChallengeParserRULE_quoted_pair   = 11
	ChallengeParserRULE_token         = 12
	ChallengeParserRULE_tchar         = 13
	ChallengeParserRULE_vchar         = 14
	ChallengeParserRULE_obs_text      = 15
)

// IHeaderContext is an interface to support dynamic dispatch.
type IHeaderContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllChallenge() []IChallengeContext
	Challenge(i int) IChallengeContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	AllSP() []antlr.TerminalNode
	SP(i int) antlr.TerminalNode
	AllHTAB() []antlr.TerminalNode
	HTAB(i int) antlr.TerminalNode

	// IsHeaderContext differentiates from other interfaces.
	IsHeaderContext()
}

type HeaderContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyHeaderContext() *HeaderContext {
	var p = new(HeaderContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_header
	return p
}

func InitEmptyHeaderContext(p *HeaderContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_header
}

func (*HeaderContext) IsHeaderContext() {}

func NewHeaderContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *HeaderContext {
	var p = new(HeaderContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_header

	return p
}

func (s *HeaderContext) GetParser() antlr.Parser { return s.parser }

func (s *HeaderContext) AllChallenge() []IChallengeContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IChallengeContext); ok {
			len++
		}
	}

	tst := make([]IChallengeContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IChallengeContext); ok {
			tst[i] = t.(IChallengeContext)
			i++
		}
	}

	return tst
}

func (s *HeaderContext) Challenge(i int) IChallengeContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IChallengeContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IChallengeContext)
}

func (s *HeaderContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserCOMMA)
}

func (s *HeaderContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserCOMMA, i)
}

func (s *HeaderContext) AllSP() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserSP)
}

func (s *HeaderContext) SP(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserSP, i)
}

func (s *HeaderContext) AllHTAB() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserHTAB)
}

func (s *HeaderContext) HTAB(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserHTAB, i)
}

func (s *HeaderContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *HeaderContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *HeaderContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterHeader(s)
	}
}

func (s *HeaderContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitHeader(s)
	}
}

func (p *ChallengeParser) Header() (localctx IHeaderContext) {
	localctx = NewHeaderContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, ChallengeParserRULE_header)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(32)
		p.Challenge()
	}
	p.SetState(49)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&16390) != 0 {
		p.SetState(36)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == ChallengeParserHTAB || _la == ChallengeParserSP {
			{
				p.SetState(33)
				_la = p.GetTokenStream().LA(1)

				if !(_la == ChallengeParserHTAB || _la == ChallengeParserSP) {
					p.GetErrorHandler().RecoverInline(p)
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
			}

			p.SetState(38)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(39)
			p.Match(ChallengeParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(43)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == ChallengeParserHTAB || _la == ChallengeParserSP {
			{
				p.SetState(40)
				_la = p.GetTokenStream().LA(1)

				if !(_la == ChallengeParserHTAB || _la == ChallengeParserSP) {
					p.GetErrorHandler().RecoverInline(p)
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
			}

			p.SetState(45)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(46)
			p.Challenge()
		}

		p.SetState(51)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IChallengeContext is an interface to support dynamic dispatch.
type IChallengeContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Auth_scheme() IAuth_schemeContext
	AllSP() []antlr.TerminalNode
	SP(i int) antlr.TerminalNode
	AllToken68() []IToken68Context
	Token68(i int) IToken68Context
	AllAuth_params() []IAuth_paramsContext
	Auth_params(i int) IAuth_paramsContext

	// IsChallengeContext differentiates from other interfaces.
	IsChallengeContext()
}

type ChallengeContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyChallengeContext() *ChallengeContext {
	var p = new(ChallengeContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_challenge
	return p
}

func InitEmptyChallengeContext(p *ChallengeContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_challenge
}

func (*ChallengeContext) IsChallengeContext() {}

func NewChallengeContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ChallengeContext {
	var p = new(ChallengeContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_challenge

	return p
}

func (s *ChallengeContext) GetParser() antlr.Parser { return s.parser }

func (s *ChallengeContext) Auth_scheme() IAuth_schemeContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAuth_schemeContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAuth_schemeContext)
}

func (s *ChallengeContext) AllSP() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserSP)
}

func (s *ChallengeContext) SP(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserSP, i)
}

func (s *ChallengeContext) AllToken68() []IToken68Context {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IToken68Context); ok {
			len++
		}
	}

	tst := make([]IToken68Context, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IToken68Context); ok {
			tst[i] = t.(IToken68Context)
			i++
		}
	}

	return tst
}

func (s *ChallengeContext) Token68(i int) IToken68Context {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IToken68Context); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IToken68Context)
}

func (s *ChallengeContext) AllAuth_params() []IAuth_paramsContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IAuth_paramsContext); ok {
			len++
		}
	}

	tst := make([]IAuth_paramsContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IAuth_paramsContext); ok {
			tst[i] = t.(IAuth_paramsContext)
			i++
		}
	}

	return tst
}

func (s *ChallengeContext) Auth_params(i int) IAuth_paramsContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAuth_paramsContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAuth_paramsContext)
}

func (s *ChallengeContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ChallengeContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ChallengeContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterChallenge(s)
	}
}

func (s *ChallengeContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitChallenge(s)
	}
}

func (p *ChallengeParser) Challenge() (localctx IChallengeContext) {
	localctx = NewChallengeContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, ChallengeParserRULE_challenge)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(52)
		p.Auth_scheme()
	}
	p.SetState(62)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(53)
				p.Match(ChallengeParserSP)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			p.SetState(58)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext()) {
			case 1:
				{
					p.SetState(54)
					p.Token68()
				}

			case 2:
				p.SetState(56)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)

				if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&93483021288) != 0 {
					{
						p.SetState(55)
						p.Auth_params()
					}

				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(64)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAuth_schemeContext is an interface to support dynamic dispatch.
type IAuth_schemeContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Token() ITokenContext

	// IsAuth_schemeContext differentiates from other interfaces.
	IsAuth_schemeContext()
}

type Auth_schemeContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAuth_schemeContext() *Auth_schemeContext {
	var p = new(Auth_schemeContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_auth_scheme
	return p
}

func InitEmptyAuth_schemeContext(p *Auth_schemeContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_auth_scheme
}

func (*Auth_schemeContext) IsAuth_schemeContext() {}

func NewAuth_schemeContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Auth_schemeContext {
	var p = new(Auth_schemeContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_auth_scheme

	return p
}

func (s *Auth_schemeContext) GetParser() antlr.Parser { return s.parser }

func (s *Auth_schemeContext) Token() ITokenContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITokenContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITokenContext)
}

func (s *Auth_schemeContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Auth_schemeContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Auth_schemeContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterAuth_scheme(s)
	}
}

func (s *Auth_schemeContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitAuth_scheme(s)
	}
}

func (p *ChallengeParser) Auth_scheme() (localctx IAuth_schemeContext) {
	localctx = NewAuth_schemeContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, ChallengeParserRULE_auth_scheme)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(65)
		p.Token()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAuth_paramsContext is an interface to support dynamic dispatch.
type IAuth_paramsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllAuth_param() []IAuth_paramContext
	Auth_param(i int) IAuth_paramContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	AllSP() []antlr.TerminalNode
	SP(i int) antlr.TerminalNode
	AllHTAB() []antlr.TerminalNode
	HTAB(i int) antlr.TerminalNode

	// IsAuth_paramsContext differentiates from other interfaces.
	IsAuth_paramsContext()
}

type Auth_paramsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAuth_paramsContext() *Auth_paramsContext {
	var p = new(Auth_paramsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_auth_params
	return p
}

func InitEmptyAuth_paramsContext(p *Auth_paramsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_auth_params
}

func (*Auth_paramsContext) IsAuth_paramsContext() {}

func NewAuth_paramsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Auth_paramsContext {
	var p = new(Auth_paramsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_auth_params

	return p
}

func (s *Auth_paramsContext) GetParser() antlr.Parser { return s.parser }

func (s *Auth_paramsContext) AllAuth_param() []IAuth_paramContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IAuth_paramContext); ok {
			len++
		}
	}

	tst := make([]IAuth_paramContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IAuth_paramContext); ok {
			tst[i] = t.(IAuth_paramContext)
			i++
		}
	}

	return tst
}

func (s *Auth_paramsContext) Auth_param(i int) IAuth_paramContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAuth_paramContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAuth_paramContext)
}

func (s *Auth_paramsContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserCOMMA)
}

func (s *Auth_paramsContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserCOMMA, i)
}

func (s *Auth_paramsContext) AllSP() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserSP)
}

func (s *Auth_paramsContext) SP(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserSP, i)
}

func (s *Auth_paramsContext) AllHTAB() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserHTAB)
}

func (s *Auth_paramsContext) HTAB(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserHTAB, i)
}

func (s *Auth_paramsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Auth_paramsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Auth_paramsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterAuth_params(s)
	}
}

func (s *Auth_paramsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitAuth_params(s)
	}
}

func (p *ChallengeParser) Auth_params() (localctx IAuth_paramsContext) {
	localctx = NewAuth_paramsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, ChallengeParserRULE_auth_params)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(67)
		p.Auth_param()
	}
	p.SetState(84)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 8, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			p.SetState(71)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == ChallengeParserHTAB || _la == ChallengeParserSP {
				{
					p.SetState(68)
					_la = p.GetTokenStream().LA(1)

					if !(_la == ChallengeParserHTAB || _la == ChallengeParserSP) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}

				p.SetState(73)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}
			{
				p.SetState(74)
				p.Match(ChallengeParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			p.SetState(78)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == ChallengeParserHTAB || _la == ChallengeParserSP {
				{
					p.SetState(75)
					_la = p.GetTokenStream().LA(1)

					if !(_la == ChallengeParserHTAB || _la == ChallengeParserSP) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}

				p.SetState(80)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}
			{
				p.SetState(81)
				p.Auth_param()
			}

		}
		p.SetState(86)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 8, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IToken68Context is an interface to support dynamic dispatch.
type IToken68Context interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllEQUALS() []antlr.TerminalNode
	EQUALS(i int) antlr.TerminalNode
	AllALPHA() []antlr.TerminalNode
	ALPHA(i int) antlr.TerminalNode
	AllDIGIT() []antlr.TerminalNode
	DIGIT(i int) antlr.TerminalNode
	AllMINUS() []antlr.TerminalNode
	MINUS(i int) antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	AllUNDERSCORE() []antlr.TerminalNode
	UNDERSCORE(i int) antlr.TerminalNode
	AllTILDE() []antlr.TerminalNode
	TILDE(i int) antlr.TerminalNode
	AllPLUS() []antlr.TerminalNode
	PLUS(i int) antlr.TerminalNode
	AllSLASH() []antlr.TerminalNode
	SLASH(i int) antlr.TerminalNode

	// IsToken68Context differentiates from other interfaces.
	IsToken68Context()
}

type Token68Context struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyToken68Context() *Token68Context {
	var p = new(Token68Context)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_token68
	return p
}

func InitEmptyToken68Context(p *Token68Context) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_token68
}

func (*Token68Context) IsToken68Context() {}

func NewToken68Context(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Token68Context {
	var p = new(Token68Context)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_token68

	return p
}

func (s *Token68Context) GetParser() antlr.Parser { return s.parser }

func (s *Token68Context) AllEQUALS() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserEQUALS)
}

func (s *Token68Context) EQUALS(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserEQUALS, i)
}

func (s *Token68Context) AllALPHA() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserALPHA)
}

func (s *Token68Context) ALPHA(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserALPHA, i)
}

func (s *Token68Context) AllDIGIT() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserDIGIT)
}

func (s *Token68Context) DIGIT(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserDIGIT, i)
}

func (s *Token68Context) AllMINUS() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserMINUS)
}

func (s *Token68Context) MINUS(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserMINUS, i)
}

func (s *Token68Context) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserCOMMA)
}

func (s *Token68Context) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserCOMMA, i)
}

func (s *Token68Context) AllUNDERSCORE() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserUNDERSCORE)
}

func (s *Token68Context) UNDERSCORE(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserUNDERSCORE, i)
}

func (s *Token68Context) AllTILDE() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserTILDE)
}

func (s *Token68Context) TILDE(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserTILDE, i)
}

func (s *Token68Context) AllPLUS() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserPLUS)
}

func (s *Token68Context) PLUS(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserPLUS, i)
}

func (s *Token68Context) AllSLASH() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserSLASH)
}

func (s *Token68Context) SLASH(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserSLASH, i)
}

func (s *Token68Context) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Token68Context) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Token68Context) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterToken68(s)
	}
}

func (s *Token68Context) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitToken68(s)
	}
}

func (p *ChallengeParser) Token68() (localctx IToken68Context) {
	localctx = NewToken68Context(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, ChallengeParserRULE_token68)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(88)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = 1
	for ok := true; ok; ok = _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		switch _alt {
		case 1:
			{
				p.SetState(87)
				_la = p.GetTokenStream().LA(1)

				if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&70934519808) != 0) {
					p.GetErrorHandler().RecoverInline(p)
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(90)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 9, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(95)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == ChallengeParserEQUALS {
		{
			p.SetState(92)
			p.Match(ChallengeParserEQUALS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(97)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAuth_paramContext is an interface to support dynamic dispatch.
type IAuth_paramContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Auth_lhs() IAuth_lhsContext
	EQUALS() antlr.TerminalNode
	Auth_rhs() IAuth_rhsContext
	AllSP() []antlr.TerminalNode
	SP(i int) antlr.TerminalNode
	AllHTAB() []antlr.TerminalNode
	HTAB(i int) antlr.TerminalNode

	// IsAuth_paramContext differentiates from other interfaces.
	IsAuth_paramContext()
}

type Auth_paramContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAuth_paramContext() *Auth_paramContext {
	var p = new(Auth_paramContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_auth_param
	return p
}

func InitEmptyAuth_paramContext(p *Auth_paramContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_auth_param
}

func (*Auth_paramContext) IsAuth_paramContext() {}

func NewAuth_paramContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Auth_paramContext {
	var p = new(Auth_paramContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_auth_param

	return p
}

func (s *Auth_paramContext) GetParser() antlr.Parser { return s.parser }

func (s *Auth_paramContext) Auth_lhs() IAuth_lhsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAuth_lhsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAuth_lhsContext)
}

func (s *Auth_paramContext) EQUALS() antlr.TerminalNode {
	return s.GetToken(ChallengeParserEQUALS, 0)
}

func (s *Auth_paramContext) Auth_rhs() IAuth_rhsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAuth_rhsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAuth_rhsContext)
}

func (s *Auth_paramContext) AllSP() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserSP)
}

func (s *Auth_paramContext) SP(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserSP, i)
}

func (s *Auth_paramContext) AllHTAB() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserHTAB)
}

func (s *Auth_paramContext) HTAB(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserHTAB, i)
}

func (s *Auth_paramContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Auth_paramContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Auth_paramContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterAuth_param(s)
	}
}

func (s *Auth_paramContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitAuth_param(s)
	}
}

func (p *ChallengeParser) Auth_param() (localctx IAuth_paramContext) {
	localctx = NewAuth_paramContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, ChallengeParserRULE_auth_param)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(98)
		p.Auth_lhs()
	}
	p.SetState(102)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == ChallengeParserHTAB || _la == ChallengeParserSP {
		{
			p.SetState(99)
			_la = p.GetTokenStream().LA(1)

			if !(_la == ChallengeParserHTAB || _la == ChallengeParserSP) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

		p.SetState(104)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(105)
		p.Match(ChallengeParserEQUALS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(109)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == ChallengeParserHTAB || _la == ChallengeParserSP {
		{
			p.SetState(106)
			_la = p.GetTokenStream().LA(1)

			if !(_la == ChallengeParserHTAB || _la == ChallengeParserSP) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

		p.SetState(111)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

	{
		p.SetState(112)
		p.Auth_rhs()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAuth_lhsContext is an interface to support dynamic dispatch.
type IAuth_lhsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Token() ITokenContext

	// IsAuth_lhsContext differentiates from other interfaces.
	IsAuth_lhsContext()
}

type Auth_lhsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAuth_lhsContext() *Auth_lhsContext {
	var p = new(Auth_lhsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_auth_lhs
	return p
}

func InitEmptyAuth_lhsContext(p *Auth_lhsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_auth_lhs
}

func (*Auth_lhsContext) IsAuth_lhsContext() {}

func NewAuth_lhsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Auth_lhsContext {
	var p = new(Auth_lhsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_auth_lhs

	return p
}

func (s *Auth_lhsContext) GetParser() antlr.Parser { return s.parser }

func (s *Auth_lhsContext) Token() ITokenContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITokenContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITokenContext)
}

func (s *Auth_lhsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Auth_lhsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Auth_lhsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterAuth_lhs(s)
	}
}

func (s *Auth_lhsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitAuth_lhs(s)
	}
}

func (p *ChallengeParser) Auth_lhs() (localctx IAuth_lhsContext) {
	localctx = NewAuth_lhsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, ChallengeParserRULE_auth_lhs)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(114)
		p.Token()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAuth_rhsContext is an interface to support dynamic dispatch.
type IAuth_rhsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Token() ITokenContext
	Quoted_string() IQuoted_stringContext

	// IsAuth_rhsContext differentiates from other interfaces.
	IsAuth_rhsContext()
}

type Auth_rhsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAuth_rhsContext() *Auth_rhsContext {
	var p = new(Auth_rhsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_auth_rhs
	return p
}

func InitEmptyAuth_rhsContext(p *Auth_rhsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_auth_rhs
}

func (*Auth_rhsContext) IsAuth_rhsContext() {}

func NewAuth_rhsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Auth_rhsContext {
	var p = new(Auth_rhsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_auth_rhs

	return p
}

func (s *Auth_rhsContext) GetParser() antlr.Parser { return s.parser }

func (s *Auth_rhsContext) Token() ITokenContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITokenContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITokenContext)
}

func (s *Auth_rhsContext) Quoted_string() IQuoted_stringContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IQuoted_stringContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IQuoted_stringContext)
}

func (s *Auth_rhsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Auth_rhsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Auth_rhsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterAuth_rhs(s)
	}
}

func (s *Auth_rhsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitAuth_rhs(s)
	}
}

func (p *ChallengeParser) Auth_rhs() (localctx IAuth_rhsContext) {
	localctx = NewAuth_rhsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, ChallengeParserRULE_auth_rhs)
	p.SetState(118)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case ChallengeParserEXCLAMATION_MARK, ChallengeParserHASH, ChallengeParserDOLLAR, ChallengeParserPERCENT, ChallengeParserAMPERSAND, ChallengeParserSQUOTE, ChallengeParserASTERISK, ChallengeParserPLUS, ChallengeParserMINUS, ChallengeParserPERIOD, ChallengeParserDIGIT, ChallengeParserALPHA, ChallengeParserCARET, ChallengeParserUNDERSCORE, ChallengeParserGRAVE, ChallengeParserPIPE, ChallengeParserTILDE:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(116)
			p.Token()
		}

	case ChallengeParserDQUOTE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(117)
			p.Quoted_string()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IRwsContext is an interface to support dynamic dispatch.
type IRwsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllSP() []antlr.TerminalNode
	SP(i int) antlr.TerminalNode
	AllHTAB() []antlr.TerminalNode
	HTAB(i int) antlr.TerminalNode

	// IsRwsContext differentiates from other interfaces.
	IsRwsContext()
}

type RwsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRwsContext() *RwsContext {
	var p = new(RwsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_rws
	return p
}

func InitEmptyRwsContext(p *RwsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_rws
}

func (*RwsContext) IsRwsContext() {}

func NewRwsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RwsContext {
	var p = new(RwsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_rws

	return p
}

func (s *RwsContext) GetParser() antlr.Parser { return s.parser }

func (s *RwsContext) AllSP() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserSP)
}

func (s *RwsContext) SP(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserSP, i)
}

func (s *RwsContext) AllHTAB() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserHTAB)
}

func (s *RwsContext) HTAB(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserHTAB, i)
}

func (s *RwsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RwsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RwsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterRws(s)
	}
}

func (s *RwsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitRws(s)
	}
}

func (p *ChallengeParser) Rws() (localctx IRwsContext) {
	localctx = NewRwsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, ChallengeParserRULE_rws)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(121)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == ChallengeParserHTAB || _la == ChallengeParserSP {
		{
			p.SetState(120)
			_la = p.GetTokenStream().LA(1)

			if !(_la == ChallengeParserHTAB || _la == ChallengeParserSP) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

		p.SetState(123)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IQuoted_stringContext is an interface to support dynamic dispatch.
type IQuoted_stringContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllDQUOTE() []antlr.TerminalNode
	DQUOTE(i int) antlr.TerminalNode
	AllQd_text() []IQd_textContext
	Qd_text(i int) IQd_textContext
	AllQuoted_pair() []IQuoted_pairContext
	Quoted_pair(i int) IQuoted_pairContext

	// IsQuoted_stringContext differentiates from other interfaces.
	IsQuoted_stringContext()
}

type Quoted_stringContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyQuoted_stringContext() *Quoted_stringContext {
	var p = new(Quoted_stringContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_quoted_string
	return p
}

func InitEmptyQuoted_stringContext(p *Quoted_stringContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_quoted_string
}

func (*Quoted_stringContext) IsQuoted_stringContext() {}

func NewQuoted_stringContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Quoted_stringContext {
	var p = new(Quoted_stringContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_quoted_string

	return p
}

func (s *Quoted_stringContext) GetParser() antlr.Parser { return s.parser }

func (s *Quoted_stringContext) AllDQUOTE() []antlr.TerminalNode {
	return s.GetTokens(ChallengeParserDQUOTE)
}

func (s *Quoted_stringContext) DQUOTE(i int) antlr.TerminalNode {
	return s.GetToken(ChallengeParserDQUOTE, i)
}

func (s *Quoted_stringContext) AllQd_text() []IQd_textContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IQd_textContext); ok {
			len++
		}
	}

	tst := make([]IQd_textContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IQd_textContext); ok {
			tst[i] = t.(IQd_textContext)
			i++
		}
	}

	return tst
}

func (s *Quoted_stringContext) Qd_text(i int) IQd_textContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IQd_textContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IQd_textContext)
}

func (s *Quoted_stringContext) AllQuoted_pair() []IQuoted_pairContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IQuoted_pairContext); ok {
			len++
		}
	}

	tst := make([]IQuoted_pairContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IQuoted_pairContext); ok {
			tst[i] = t.(IQuoted_pairContext)
			i++
		}
	}

	return tst
}

func (s *Quoted_stringContext) Quoted_pair(i int) IQuoted_pairContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IQuoted_pairContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IQuoted_pairContext)
}

func (s *Quoted_stringContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Quoted_stringContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Quoted_stringContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterQuoted_string(s)
	}
}

func (s *Quoted_stringContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitQuoted_string(s)
	}
}

func (p *ChallengeParser) Quoted_string() (localctx IQuoted_stringContext) {
	localctx = NewQuoted_stringContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, ChallengeParserRULE_quoted_string)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(125)
		p.Match(ChallengeParserDQUOTE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(128)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&274877906926) != 0) {
		p.SetState(128)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case ChallengeParserHTAB, ChallengeParserSP, ChallengeParserEXCLAMATION_MARK, ChallengeParserHASH, ChallengeParserDOLLAR, ChallengeParserPERCENT, ChallengeParserAMPERSAND, ChallengeParserSQUOTE, ChallengeParserOPEN_PARENS, ChallengeParserCLOSE_PARENS, ChallengeParserASTERISK, ChallengeParserPLUS, ChallengeParserCOMMA, ChallengeParserMINUS, ChallengeParserPERIOD, ChallengeParserSLASH, ChallengeParserDIGIT, ChallengeParserCOLON, ChallengeParserSEMICOLON, ChallengeParserLESS_THAN, ChallengeParserEQUALS, ChallengeParserGREATER_THAN, ChallengeParserQUESTION, ChallengeParserAT, ChallengeParserALPHA, ChallengeParserOPEN_BRACKET, ChallengeParserCLOSE_BRACKET, ChallengeParserCARET, ChallengeParserUNDERSCORE, ChallengeParserGRAVE, ChallengeParserOPEN_BRACE, ChallengeParserPIPE, ChallengeParserCLOSE_BRACE, ChallengeParserTILDE, ChallengeParserEXTENDED_ASCII:
			{
				p.SetState(126)
				p.Qd_text()
			}

		case ChallengeParserBACKSLASH:
			{
				p.SetState(127)
				p.Quoted_pair()
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(130)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(132)
		p.Match(ChallengeParserDQUOTE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IQd_textContext is an interface to support dynamic dispatch.
type IQd_textContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	HTAB() antlr.TerminalNode
	SP() antlr.TerminalNode
	EXCLAMATION_MARK() antlr.TerminalNode
	HASH() antlr.TerminalNode
	DOLLAR() antlr.TerminalNode
	PERCENT() antlr.TerminalNode
	AMPERSAND() antlr.TerminalNode
	SQUOTE() antlr.TerminalNode
	OPEN_PARENS() antlr.TerminalNode
	CLOSE_PARENS() antlr.TerminalNode
	ASTERISK() antlr.TerminalNode
	PLUS() antlr.TerminalNode
	COMMA() antlr.TerminalNode
	MINUS() antlr.TerminalNode
	PERIOD() antlr.TerminalNode
	SLASH() antlr.TerminalNode
	DIGIT() antlr.TerminalNode
	COLON() antlr.TerminalNode
	SEMICOLON() antlr.TerminalNode
	LESS_THAN() antlr.TerminalNode
	EQUALS() antlr.TerminalNode
	GREATER_THAN() antlr.TerminalNode
	QUESTION() antlr.TerminalNode
	AT() antlr.TerminalNode
	OPEN_BRACKET() antlr.TerminalNode
	CLOSE_BRACKET() antlr.TerminalNode
	CARET() antlr.TerminalNode
	UNDERSCORE() antlr.TerminalNode
	GRAVE() antlr.TerminalNode
	ALPHA() antlr.TerminalNode
	OPEN_BRACE() antlr.TerminalNode
	PIPE() antlr.TerminalNode
	CLOSE_BRACE() antlr.TerminalNode
	TILDE() antlr.TerminalNode
	Obs_text() IObs_textContext

	// IsQd_textContext differentiates from other interfaces.
	IsQd_textContext()
}

type Qd_textContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyQd_textContext() *Qd_textContext {
	var p = new(Qd_textContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_qd_text
	return p
}

func InitEmptyQd_textContext(p *Qd_textContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_qd_text
}

func (*Qd_textContext) IsQd_textContext() {}

func NewQd_textContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Qd_textContext {
	var p = new(Qd_textContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_qd_text

	return p
}

func (s *Qd_textContext) GetParser() antlr.Parser { return s.parser }

func (s *Qd_textContext) HTAB() antlr.TerminalNode {
	return s.GetToken(ChallengeParserHTAB, 0)
}

func (s *Qd_textContext) SP() antlr.TerminalNode {
	return s.GetToken(ChallengeParserSP, 0)
}

func (s *Qd_textContext) EXCLAMATION_MARK() antlr.TerminalNode {
	return s.GetToken(ChallengeParserEXCLAMATION_MARK, 0)
}

func (s *Qd_textContext) HASH() antlr.TerminalNode {
	return s.GetToken(ChallengeParserHASH, 0)
}

func (s *Qd_textContext) DOLLAR() antlr.TerminalNode {
	return s.GetToken(ChallengeParserDOLLAR, 0)
}

func (s *Qd_textContext) PERCENT() antlr.TerminalNode {
	return s.GetToken(ChallengeParserPERCENT, 0)
}

func (s *Qd_textContext) AMPERSAND() antlr.TerminalNode {
	return s.GetToken(ChallengeParserAMPERSAND, 0)
}

func (s *Qd_textContext) SQUOTE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserSQUOTE, 0)
}

func (s *Qd_textContext) OPEN_PARENS() antlr.TerminalNode {
	return s.GetToken(ChallengeParserOPEN_PARENS, 0)
}

func (s *Qd_textContext) CLOSE_PARENS() antlr.TerminalNode {
	return s.GetToken(ChallengeParserCLOSE_PARENS, 0)
}

func (s *Qd_textContext) ASTERISK() antlr.TerminalNode {
	return s.GetToken(ChallengeParserASTERISK, 0)
}

func (s *Qd_textContext) PLUS() antlr.TerminalNode {
	return s.GetToken(ChallengeParserPLUS, 0)
}

func (s *Qd_textContext) COMMA() antlr.TerminalNode {
	return s.GetToken(ChallengeParserCOMMA, 0)
}

func (s *Qd_textContext) MINUS() antlr.TerminalNode {
	return s.GetToken(ChallengeParserMINUS, 0)
}

func (s *Qd_textContext) PERIOD() antlr.TerminalNode {
	return s.GetToken(ChallengeParserPERIOD, 0)
}

func (s *Qd_textContext) SLASH() antlr.TerminalNode {
	return s.GetToken(ChallengeParserSLASH, 0)
}

func (s *Qd_textContext) DIGIT() antlr.TerminalNode {
	return s.GetToken(ChallengeParserDIGIT, 0)
}

func (s *Qd_textContext) COLON() antlr.TerminalNode {
	return s.GetToken(ChallengeParserCOLON, 0)
}

func (s *Qd_textContext) SEMICOLON() antlr.TerminalNode {
	return s.GetToken(ChallengeParserSEMICOLON, 0)
}

func (s *Qd_textContext) LESS_THAN() antlr.TerminalNode {
	return s.GetToken(ChallengeParserLESS_THAN, 0)
}

func (s *Qd_textContext) EQUALS() antlr.TerminalNode {
	return s.GetToken(ChallengeParserEQUALS, 0)
}

func (s *Qd_textContext) GREATER_THAN() antlr.TerminalNode {
	return s.GetToken(ChallengeParserGREATER_THAN, 0)
}

func (s *Qd_textContext) QUESTION() antlr.TerminalNode {
	return s.GetToken(ChallengeParserQUESTION, 0)
}

func (s *Qd_textContext) AT() antlr.TerminalNode {
	return s.GetToken(ChallengeParserAT, 0)
}

func (s *Qd_textContext) OPEN_BRACKET() antlr.TerminalNode {
	return s.GetToken(ChallengeParserOPEN_BRACKET, 0)
}

func (s *Qd_textContext) CLOSE_BRACKET() antlr.TerminalNode {
	return s.GetToken(ChallengeParserCLOSE_BRACKET, 0)
}

func (s *Qd_textContext) CARET() antlr.TerminalNode {
	return s.GetToken(ChallengeParserCARET, 0)
}

func (s *Qd_textContext) UNDERSCORE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserUNDERSCORE, 0)
}

func (s *Qd_textContext) GRAVE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserGRAVE, 0)
}

func (s *Qd_textContext) ALPHA() antlr.TerminalNode {
	return s.GetToken(ChallengeParserALPHA, 0)
}

func (s *Qd_textContext) OPEN_BRACE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserOPEN_BRACE, 0)
}

func (s *Qd_textContext) PIPE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserPIPE, 0)
}

func (s *Qd_textContext) CLOSE_BRACE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserCLOSE_BRACE, 0)
}

func (s *Qd_textContext) TILDE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserTILDE, 0)
}

func (s *Qd_textContext) Obs_text() IObs_textContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IObs_textContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IObs_textContext)
}

func (s *Qd_textContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Qd_textContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Qd_textContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterQd_text(s)
	}
}

func (s *Qd_textContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitQd_text(s)
	}
}

func (p *ChallengeParser) Qd_text() (localctx IQd_textContext) {
	localctx = NewQd_textContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, ChallengeParserRULE_qd_text)
	p.SetState(169)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case ChallengeParserHTAB:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(134)
			p.Match(ChallengeParserHTAB)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserSP:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(135)
			p.Match(ChallengeParserSP)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserEXCLAMATION_MARK:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(136)
			p.Match(ChallengeParserEXCLAMATION_MARK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserHASH:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(137)
			p.Match(ChallengeParserHASH)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserDOLLAR:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(138)
			p.Match(ChallengeParserDOLLAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserPERCENT:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(139)
			p.Match(ChallengeParserPERCENT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserAMPERSAND:
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(140)
			p.Match(ChallengeParserAMPERSAND)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserSQUOTE:
		p.EnterOuterAlt(localctx, 8)
		{
			p.SetState(141)
			p.Match(ChallengeParserSQUOTE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserOPEN_PARENS:
		p.EnterOuterAlt(localctx, 9)
		{
			p.SetState(142)
			p.Match(ChallengeParserOPEN_PARENS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserCLOSE_PARENS:
		p.EnterOuterAlt(localctx, 10)
		{
			p.SetState(143)
			p.Match(ChallengeParserCLOSE_PARENS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserASTERISK:
		p.EnterOuterAlt(localctx, 11)
		{
			p.SetState(144)
			p.Match(ChallengeParserASTERISK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserPLUS:
		p.EnterOuterAlt(localctx, 12)
		{
			p.SetState(145)
			p.Match(ChallengeParserPLUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserCOMMA:
		p.EnterOuterAlt(localctx, 13)
		{
			p.SetState(146)
			p.Match(ChallengeParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserMINUS:
		p.EnterOuterAlt(localctx, 14)
		{
			p.SetState(147)
			p.Match(ChallengeParserMINUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserPERIOD:
		p.EnterOuterAlt(localctx, 15)
		{
			p.SetState(148)
			p.Match(ChallengeParserPERIOD)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserSLASH:
		p.EnterOuterAlt(localctx, 16)
		{
			p.SetState(149)
			p.Match(ChallengeParserSLASH)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserDIGIT:
		p.EnterOuterAlt(localctx, 17)
		{
			p.SetState(150)
			p.Match(ChallengeParserDIGIT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserCOLON:
		p.EnterOuterAlt(localctx, 18)
		{
			p.SetState(151)
			p.Match(ChallengeParserCOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserSEMICOLON:
		p.EnterOuterAlt(localctx, 19)
		{
			p.SetState(152)
			p.Match(ChallengeParserSEMICOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserLESS_THAN:
		p.EnterOuterAlt(localctx, 20)
		{
			p.SetState(153)
			p.Match(ChallengeParserLESS_THAN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserEQUALS:
		p.EnterOuterAlt(localctx, 21)
		{
			p.SetState(154)
			p.Match(ChallengeParserEQUALS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserGREATER_THAN:
		p.EnterOuterAlt(localctx, 22)
		{
			p.SetState(155)
			p.Match(ChallengeParserGREATER_THAN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserQUESTION:
		p.EnterOuterAlt(localctx, 23)
		{
			p.SetState(156)
			p.Match(ChallengeParserQUESTION)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserAT:
		p.EnterOuterAlt(localctx, 24)
		{
			p.SetState(157)
			p.Match(ChallengeParserAT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserOPEN_BRACKET:
		p.EnterOuterAlt(localctx, 25)
		{
			p.SetState(158)
			p.Match(ChallengeParserOPEN_BRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserCLOSE_BRACKET:
		p.EnterOuterAlt(localctx, 26)
		{
			p.SetState(159)
			p.Match(ChallengeParserCLOSE_BRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserCARET:
		p.EnterOuterAlt(localctx, 27)
		{
			p.SetState(160)
			p.Match(ChallengeParserCARET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserUNDERSCORE:
		p.EnterOuterAlt(localctx, 28)
		{
			p.SetState(161)
			p.Match(ChallengeParserUNDERSCORE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserGRAVE:
		p.EnterOuterAlt(localctx, 29)
		{
			p.SetState(162)
			p.Match(ChallengeParserGRAVE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserALPHA:
		p.EnterOuterAlt(localctx, 30)
		{
			p.SetState(163)
			p.Match(ChallengeParserALPHA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserOPEN_BRACE:
		p.EnterOuterAlt(localctx, 31)
		{
			p.SetState(164)
			p.Match(ChallengeParserOPEN_BRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserPIPE:
		p.EnterOuterAlt(localctx, 32)
		{
			p.SetState(165)
			p.Match(ChallengeParserPIPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserCLOSE_BRACE:
		p.EnterOuterAlt(localctx, 33)
		{
			p.SetState(166)
			p.Match(ChallengeParserCLOSE_BRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserTILDE:
		p.EnterOuterAlt(localctx, 34)
		{
			p.SetState(167)
			p.Match(ChallengeParserTILDE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ChallengeParserEXTENDED_ASCII:
		p.EnterOuterAlt(localctx, 35)
		{
			p.SetState(168)
			p.Obs_text()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IQuoted_pairContext is an interface to support dynamic dispatch.
type IQuoted_pairContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	BACKSLASH() antlr.TerminalNode
	HTAB() antlr.TerminalNode
	SP() antlr.TerminalNode
	Vchar() IVcharContext
	Obs_text() IObs_textContext

	// IsQuoted_pairContext differentiates from other interfaces.
	IsQuoted_pairContext()
}

type Quoted_pairContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyQuoted_pairContext() *Quoted_pairContext {
	var p = new(Quoted_pairContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_quoted_pair
	return p
}

func InitEmptyQuoted_pairContext(p *Quoted_pairContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_quoted_pair
}

func (*Quoted_pairContext) IsQuoted_pairContext() {}

func NewQuoted_pairContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Quoted_pairContext {
	var p = new(Quoted_pairContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_quoted_pair

	return p
}

func (s *Quoted_pairContext) GetParser() antlr.Parser { return s.parser }

func (s *Quoted_pairContext) BACKSLASH() antlr.TerminalNode {
	return s.GetToken(ChallengeParserBACKSLASH, 0)
}

func (s *Quoted_pairContext) HTAB() antlr.TerminalNode {
	return s.GetToken(ChallengeParserHTAB, 0)
}

func (s *Quoted_pairContext) SP() antlr.TerminalNode {
	return s.GetToken(ChallengeParserSP, 0)
}

func (s *Quoted_pairContext) Vchar() IVcharContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVcharContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IVcharContext)
}

func (s *Quoted_pairContext) Obs_text() IObs_textContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IObs_textContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IObs_textContext)
}

func (s *Quoted_pairContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Quoted_pairContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Quoted_pairContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterQuoted_pair(s)
	}
}

func (s *Quoted_pairContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitQuoted_pair(s)
	}
}

func (p *ChallengeParser) Quoted_pair() (localctx IQuoted_pairContext) {
	localctx = NewQuoted_pairContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, ChallengeParserRULE_quoted_pair)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(171)
		p.Match(ChallengeParserBACKSLASH)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(176)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 18, p.GetParserRuleContext()) {
	case 1:
		{
			p.SetState(172)
			p.Match(ChallengeParserHTAB)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		{
			p.SetState(173)
			p.Match(ChallengeParserSP)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		{
			p.SetState(174)
			p.Vchar()
		}

	case 4:
		{
			p.SetState(175)
			p.Obs_text()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITokenContext is an interface to support dynamic dispatch.
type ITokenContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllTchar() []ITcharContext
	Tchar(i int) ITcharContext

	// IsTokenContext differentiates from other interfaces.
	IsTokenContext()
}

type TokenContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTokenContext() *TokenContext {
	var p = new(TokenContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_token
	return p
}

func InitEmptyTokenContext(p *TokenContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_token
}

func (*TokenContext) IsTokenContext() {}

func NewTokenContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TokenContext {
	var p = new(TokenContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_token

	return p
}

func (s *TokenContext) GetParser() antlr.Parser { return s.parser }

func (s *TokenContext) AllTchar() []ITcharContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITcharContext); ok {
			len++
		}
	}

	tst := make([]ITcharContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITcharContext); ok {
			tst[i] = t.(ITcharContext)
			i++
		}
	}

	return tst
}

func (s *TokenContext) Tchar(i int) ITcharContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITcharContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITcharContext)
}

func (s *TokenContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TokenContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TokenContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterToken(s)
	}
}

func (s *TokenContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitToken(s)
	}
}

func (p *ChallengeParser) Token() (localctx ITokenContext) {
	localctx = NewTokenContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, ChallengeParserRULE_token)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(179)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&93483021288) != 0) {
		{
			p.SetState(178)
			p.Tchar()
		}

		p.SetState(181)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITcharContext is an interface to support dynamic dispatch.
type ITcharContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EXCLAMATION_MARK() antlr.TerminalNode
	HASH() antlr.TerminalNode
	DOLLAR() antlr.TerminalNode
	PERCENT() antlr.TerminalNode
	AMPERSAND() antlr.TerminalNode
	SQUOTE() antlr.TerminalNode
	ASTERISK() antlr.TerminalNode
	PLUS() antlr.TerminalNode
	MINUS() antlr.TerminalNode
	PERIOD() antlr.TerminalNode
	CARET() antlr.TerminalNode
	UNDERSCORE() antlr.TerminalNode
	GRAVE() antlr.TerminalNode
	PIPE() antlr.TerminalNode
	TILDE() antlr.TerminalNode
	DIGIT() antlr.TerminalNode
	ALPHA() antlr.TerminalNode

	// IsTcharContext differentiates from other interfaces.
	IsTcharContext()
}

type TcharContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTcharContext() *TcharContext {
	var p = new(TcharContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_tchar
	return p
}

func InitEmptyTcharContext(p *TcharContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_tchar
}

func (*TcharContext) IsTcharContext() {}

func NewTcharContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TcharContext {
	var p = new(TcharContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_tchar

	return p
}

func (s *TcharContext) GetParser() antlr.Parser { return s.parser }

func (s *TcharContext) EXCLAMATION_MARK() antlr.TerminalNode {
	return s.GetToken(ChallengeParserEXCLAMATION_MARK, 0)
}

func (s *TcharContext) HASH() antlr.TerminalNode {
	return s.GetToken(ChallengeParserHASH, 0)
}

func (s *TcharContext) DOLLAR() antlr.TerminalNode {
	return s.GetToken(ChallengeParserDOLLAR, 0)
}

func (s *TcharContext) PERCENT() antlr.TerminalNode {
	return s.GetToken(ChallengeParserPERCENT, 0)
}

func (s *TcharContext) AMPERSAND() antlr.TerminalNode {
	return s.GetToken(ChallengeParserAMPERSAND, 0)
}

func (s *TcharContext) SQUOTE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserSQUOTE, 0)
}

func (s *TcharContext) ASTERISK() antlr.TerminalNode {
	return s.GetToken(ChallengeParserASTERISK, 0)
}

func (s *TcharContext) PLUS() antlr.TerminalNode {
	return s.GetToken(ChallengeParserPLUS, 0)
}

func (s *TcharContext) MINUS() antlr.TerminalNode {
	return s.GetToken(ChallengeParserMINUS, 0)
}

func (s *TcharContext) PERIOD() antlr.TerminalNode {
	return s.GetToken(ChallengeParserPERIOD, 0)
}

func (s *TcharContext) CARET() antlr.TerminalNode {
	return s.GetToken(ChallengeParserCARET, 0)
}

func (s *TcharContext) UNDERSCORE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserUNDERSCORE, 0)
}

func (s *TcharContext) GRAVE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserGRAVE, 0)
}

func (s *TcharContext) PIPE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserPIPE, 0)
}

func (s *TcharContext) TILDE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserTILDE, 0)
}

func (s *TcharContext) DIGIT() antlr.TerminalNode {
	return s.GetToken(ChallengeParserDIGIT, 0)
}

func (s *TcharContext) ALPHA() antlr.TerminalNode {
	return s.GetToken(ChallengeParserALPHA, 0)
}

func (s *TcharContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TcharContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TcharContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterTchar(s)
	}
}

func (s *TcharContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitTchar(s)
	}
}

func (p *ChallengeParser) Tchar() (localctx ITcharContext) {
	localctx = NewTcharContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, ChallengeParserRULE_tchar)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(183)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&93483021288) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IVcharContext is an interface to support dynamic dispatch.
type IVcharContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EXCLAMATION_MARK() antlr.TerminalNode
	DQUOTE() antlr.TerminalNode
	HASH() antlr.TerminalNode
	DOLLAR() antlr.TerminalNode
	PERCENT() antlr.TerminalNode
	AMPERSAND() antlr.TerminalNode
	SQUOTE() antlr.TerminalNode
	OPEN_PARENS() antlr.TerminalNode
	CLOSE_PARENS() antlr.TerminalNode
	ASTERISK() antlr.TerminalNode
	PLUS() antlr.TerminalNode
	COMMA() antlr.TerminalNode
	MINUS() antlr.TerminalNode
	PERIOD() antlr.TerminalNode
	SLASH() antlr.TerminalNode
	DIGIT() antlr.TerminalNode
	COLON() antlr.TerminalNode
	SEMICOLON() antlr.TerminalNode
	LESS_THAN() antlr.TerminalNode
	EQUALS() antlr.TerminalNode
	GREATER_THAN() antlr.TerminalNode
	QUESTION() antlr.TerminalNode
	AT() antlr.TerminalNode
	OPEN_BRACKET() antlr.TerminalNode
	BACKSLASH() antlr.TerminalNode
	CLOSE_BRACKET() antlr.TerminalNode
	CARET() antlr.TerminalNode
	UNDERSCORE() antlr.TerminalNode
	GRAVE() antlr.TerminalNode
	ALPHA() antlr.TerminalNode
	OPEN_BRACE() antlr.TerminalNode
	PIPE() antlr.TerminalNode
	CLOSE_BRACE() antlr.TerminalNode
	TILDE() antlr.TerminalNode
	EXTENDED_ASCII() antlr.TerminalNode

	// IsVcharContext differentiates from other interfaces.
	IsVcharContext()
}

type VcharContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyVcharContext() *VcharContext {
	var p = new(VcharContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_vchar
	return p
}

func InitEmptyVcharContext(p *VcharContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_vchar
}

func (*VcharContext) IsVcharContext() {}

func NewVcharContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *VcharContext {
	var p = new(VcharContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_vchar

	return p
}

func (s *VcharContext) GetParser() antlr.Parser { return s.parser }

func (s *VcharContext) EXCLAMATION_MARK() antlr.TerminalNode {
	return s.GetToken(ChallengeParserEXCLAMATION_MARK, 0)
}

func (s *VcharContext) DQUOTE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserDQUOTE, 0)
}

func (s *VcharContext) HASH() antlr.TerminalNode {
	return s.GetToken(ChallengeParserHASH, 0)
}

func (s *VcharContext) DOLLAR() antlr.TerminalNode {
	return s.GetToken(ChallengeParserDOLLAR, 0)
}

func (s *VcharContext) PERCENT() antlr.TerminalNode {
	return s.GetToken(ChallengeParserPERCENT, 0)
}

func (s *VcharContext) AMPERSAND() antlr.TerminalNode {
	return s.GetToken(ChallengeParserAMPERSAND, 0)
}

func (s *VcharContext) SQUOTE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserSQUOTE, 0)
}

func (s *VcharContext) OPEN_PARENS() antlr.TerminalNode {
	return s.GetToken(ChallengeParserOPEN_PARENS, 0)
}

func (s *VcharContext) CLOSE_PARENS() antlr.TerminalNode {
	return s.GetToken(ChallengeParserCLOSE_PARENS, 0)
}

func (s *VcharContext) ASTERISK() antlr.TerminalNode {
	return s.GetToken(ChallengeParserASTERISK, 0)
}

func (s *VcharContext) PLUS() antlr.TerminalNode {
	return s.GetToken(ChallengeParserPLUS, 0)
}

func (s *VcharContext) COMMA() antlr.TerminalNode {
	return s.GetToken(ChallengeParserCOMMA, 0)
}

func (s *VcharContext) MINUS() antlr.TerminalNode {
	return s.GetToken(ChallengeParserMINUS, 0)
}

func (s *VcharContext) PERIOD() antlr.TerminalNode {
	return s.GetToken(ChallengeParserPERIOD, 0)
}

func (s *VcharContext) SLASH() antlr.TerminalNode {
	return s.GetToken(ChallengeParserSLASH, 0)
}

func (s *VcharContext) DIGIT() antlr.TerminalNode {
	return s.GetToken(ChallengeParserDIGIT, 0)
}

func (s *VcharContext) COLON() antlr.TerminalNode {
	return s.GetToken(ChallengeParserCOLON, 0)
}

func (s *VcharContext) SEMICOLON() antlr.TerminalNode {
	return s.GetToken(ChallengeParserSEMICOLON, 0)
}

func (s *VcharContext) LESS_THAN() antlr.TerminalNode {
	return s.GetToken(ChallengeParserLESS_THAN, 0)
}

func (s *VcharContext) EQUALS() antlr.TerminalNode {
	return s.GetToken(ChallengeParserEQUALS, 0)
}

func (s *VcharContext) GREATER_THAN() antlr.TerminalNode {
	return s.GetToken(ChallengeParserGREATER_THAN, 0)
}

func (s *VcharContext) QUESTION() antlr.TerminalNode {
	return s.GetToken(ChallengeParserQUESTION, 0)
}

func (s *VcharContext) AT() antlr.TerminalNode {
	return s.GetToken(ChallengeParserAT, 0)
}

func (s *VcharContext) OPEN_BRACKET() antlr.TerminalNode {
	return s.GetToken(ChallengeParserOPEN_BRACKET, 0)
}

func (s *VcharContext) BACKSLASH() antlr.TerminalNode {
	return s.GetToken(ChallengeParserBACKSLASH, 0)
}

func (s *VcharContext) CLOSE_BRACKET() antlr.TerminalNode {
	return s.GetToken(ChallengeParserCLOSE_BRACKET, 0)
}

func (s *VcharContext) CARET() antlr.TerminalNode {
	return s.GetToken(ChallengeParserCARET, 0)
}

func (s *VcharContext) UNDERSCORE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserUNDERSCORE, 0)
}

func (s *VcharContext) GRAVE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserGRAVE, 0)
}

func (s *VcharContext) ALPHA() antlr.TerminalNode {
	return s.GetToken(ChallengeParserALPHA, 0)
}

func (s *VcharContext) OPEN_BRACE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserOPEN_BRACE, 0)
}

func (s *VcharContext) PIPE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserPIPE, 0)
}

func (s *VcharContext) CLOSE_BRACE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserCLOSE_BRACE, 0)
}

func (s *VcharContext) TILDE() antlr.TerminalNode {
	return s.GetToken(ChallengeParserTILDE, 0)
}

func (s *VcharContext) EXTENDED_ASCII() antlr.TerminalNode {
	return s.GetToken(ChallengeParserEXTENDED_ASCII, 0)
}

func (s *VcharContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VcharContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *VcharContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterVchar(s)
	}
}

func (s *VcharContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitVchar(s)
	}
}

func (p *ChallengeParser) Vchar() (localctx IVcharContext) {
	localctx = NewVcharContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, ChallengeParserRULE_vchar)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(185)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&274877906936) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IObs_textContext is an interface to support dynamic dispatch.
type IObs_textContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EXTENDED_ASCII() antlr.TerminalNode

	// IsObs_textContext differentiates from other interfaces.
	IsObs_textContext()
}

type Obs_textContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyObs_textContext() *Obs_textContext {
	var p = new(Obs_textContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_obs_text
	return p
}

func InitEmptyObs_textContext(p *Obs_textContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ChallengeParserRULE_obs_text
}

func (*Obs_textContext) IsObs_textContext() {}

func NewObs_textContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Obs_textContext {
	var p = new(Obs_textContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ChallengeParserRULE_obs_text

	return p
}

func (s *Obs_textContext) GetParser() antlr.Parser { return s.parser }

func (s *Obs_textContext) EXTENDED_ASCII() antlr.TerminalNode {
	return s.GetToken(ChallengeParserEXTENDED_ASCII, 0)
}

func (s *Obs_textContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Obs_textContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Obs_textContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.EnterObs_text(s)
	}
}

func (s *Obs_textContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ChallengeListener); ok {
		listenerT.ExitObs_text(s)
	}
}

func (p *ChallengeParser) Obs_text() (localctx IObs_textContext) {
	localctx = NewObs_textContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, ChallengeParserRULE_obs_text)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(187)
		p.Match(ChallengeParserEXTENDED_ASCII)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}
