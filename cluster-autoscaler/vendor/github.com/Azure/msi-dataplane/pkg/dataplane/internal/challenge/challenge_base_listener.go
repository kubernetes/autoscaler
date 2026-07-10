// Code generated from Challenge.g4 by ANTLR 4.13.2. DO NOT EDIT.

package challenge // Challenge
import "github.com/antlr4-go/antlr/v4"

// BaseChallengeListener is a complete listener for a parse tree produced by ChallengeParser.
type BaseChallengeListener struct{}

var _ ChallengeListener = &BaseChallengeListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseChallengeListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseChallengeListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseChallengeListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseChallengeListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterHeader is called when production header is entered.
func (s *BaseChallengeListener) EnterHeader(ctx *HeaderContext) {}

// ExitHeader is called when production header is exited.
func (s *BaseChallengeListener) ExitHeader(ctx *HeaderContext) {}

// EnterChallenge is called when production challenge is entered.
func (s *BaseChallengeListener) EnterChallenge(ctx *ChallengeContext) {}

// ExitChallenge is called when production challenge is exited.
func (s *BaseChallengeListener) ExitChallenge(ctx *ChallengeContext) {}

// EnterAuth_scheme is called when production auth_scheme is entered.
func (s *BaseChallengeListener) EnterAuth_scheme(ctx *Auth_schemeContext) {}

// ExitAuth_scheme is called when production auth_scheme is exited.
func (s *BaseChallengeListener) ExitAuth_scheme(ctx *Auth_schemeContext) {}

// EnterAuth_params is called when production auth_params is entered.
func (s *BaseChallengeListener) EnterAuth_params(ctx *Auth_paramsContext) {}

// ExitAuth_params is called when production auth_params is exited.
func (s *BaseChallengeListener) ExitAuth_params(ctx *Auth_paramsContext) {}

// EnterToken68 is called when production token68 is entered.
func (s *BaseChallengeListener) EnterToken68(ctx *Token68Context) {}

// ExitToken68 is called when production token68 is exited.
func (s *BaseChallengeListener) ExitToken68(ctx *Token68Context) {}

// EnterAuth_param is called when production auth_param is entered.
func (s *BaseChallengeListener) EnterAuth_param(ctx *Auth_paramContext) {}

// ExitAuth_param is called when production auth_param is exited.
func (s *BaseChallengeListener) ExitAuth_param(ctx *Auth_paramContext) {}

// EnterAuth_lhs is called when production auth_lhs is entered.
func (s *BaseChallengeListener) EnterAuth_lhs(ctx *Auth_lhsContext) {}

// ExitAuth_lhs is called when production auth_lhs is exited.
func (s *BaseChallengeListener) ExitAuth_lhs(ctx *Auth_lhsContext) {}

// EnterAuth_rhs is called when production auth_rhs is entered.
func (s *BaseChallengeListener) EnterAuth_rhs(ctx *Auth_rhsContext) {}

// ExitAuth_rhs is called when production auth_rhs is exited.
func (s *BaseChallengeListener) ExitAuth_rhs(ctx *Auth_rhsContext) {}

// EnterRws is called when production rws is entered.
func (s *BaseChallengeListener) EnterRws(ctx *RwsContext) {}

// ExitRws is called when production rws is exited.
func (s *BaseChallengeListener) ExitRws(ctx *RwsContext) {}

// EnterQuoted_string is called when production quoted_string is entered.
func (s *BaseChallengeListener) EnterQuoted_string(ctx *Quoted_stringContext) {}

// ExitQuoted_string is called when production quoted_string is exited.
func (s *BaseChallengeListener) ExitQuoted_string(ctx *Quoted_stringContext) {}

// EnterQd_text is called when production qd_text is entered.
func (s *BaseChallengeListener) EnterQd_text(ctx *Qd_textContext) {}

// ExitQd_text is called when production qd_text is exited.
func (s *BaseChallengeListener) ExitQd_text(ctx *Qd_textContext) {}

// EnterQuoted_pair is called when production quoted_pair is entered.
func (s *BaseChallengeListener) EnterQuoted_pair(ctx *Quoted_pairContext) {}

// ExitQuoted_pair is called when production quoted_pair is exited.
func (s *BaseChallengeListener) ExitQuoted_pair(ctx *Quoted_pairContext) {}

// EnterToken is called when production token is entered.
func (s *BaseChallengeListener) EnterToken(ctx *TokenContext) {}

// ExitToken is called when production token is exited.
func (s *BaseChallengeListener) ExitToken(ctx *TokenContext) {}

// EnterTchar is called when production tchar is entered.
func (s *BaseChallengeListener) EnterTchar(ctx *TcharContext) {}

// ExitTchar is called when production tchar is exited.
func (s *BaseChallengeListener) ExitTchar(ctx *TcharContext) {}

// EnterVchar is called when production vchar is entered.
func (s *BaseChallengeListener) EnterVchar(ctx *VcharContext) {}

// ExitVchar is called when production vchar is exited.
func (s *BaseChallengeListener) ExitVchar(ctx *VcharContext) {}

// EnterObs_text is called when production obs_text is entered.
func (s *BaseChallengeListener) EnterObs_text(ctx *Obs_textContext) {}

// ExitObs_text is called when production obs_text is exited.
func (s *BaseChallengeListener) ExitObs_text(ctx *Obs_textContext) {}
