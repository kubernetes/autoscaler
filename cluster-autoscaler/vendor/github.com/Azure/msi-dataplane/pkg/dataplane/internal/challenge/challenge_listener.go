// Code generated from Challenge.g4 by ANTLR 4.13.2. DO NOT EDIT.

package challenge // Challenge
import "github.com/antlr4-go/antlr/v4"

// ChallengeListener is a complete listener for a parse tree produced by ChallengeParser.
type ChallengeListener interface {
	antlr.ParseTreeListener

	// EnterHeader is called when entering the header production.
	EnterHeader(c *HeaderContext)

	// EnterChallenge is called when entering the challenge production.
	EnterChallenge(c *ChallengeContext)

	// EnterAuth_scheme is called when entering the auth_scheme production.
	EnterAuth_scheme(c *Auth_schemeContext)

	// EnterAuth_params is called when entering the auth_params production.
	EnterAuth_params(c *Auth_paramsContext)

	// EnterToken68 is called when entering the token68 production.
	EnterToken68(c *Token68Context)

	// EnterAuth_param is called when entering the auth_param production.
	EnterAuth_param(c *Auth_paramContext)

	// EnterAuth_lhs is called when entering the auth_lhs production.
	EnterAuth_lhs(c *Auth_lhsContext)

	// EnterAuth_rhs is called when entering the auth_rhs production.
	EnterAuth_rhs(c *Auth_rhsContext)

	// EnterRws is called when entering the rws production.
	EnterRws(c *RwsContext)

	// EnterQuoted_string is called when entering the quoted_string production.
	EnterQuoted_string(c *Quoted_stringContext)

	// EnterQd_text is called when entering the qd_text production.
	EnterQd_text(c *Qd_textContext)

	// EnterQuoted_pair is called when entering the quoted_pair production.
	EnterQuoted_pair(c *Quoted_pairContext)

	// EnterToken is called when entering the token production.
	EnterToken(c *TokenContext)

	// EnterTchar is called when entering the tchar production.
	EnterTchar(c *TcharContext)

	// EnterVchar is called when entering the vchar production.
	EnterVchar(c *VcharContext)

	// EnterObs_text is called when entering the obs_text production.
	EnterObs_text(c *Obs_textContext)

	// ExitHeader is called when exiting the header production.
	ExitHeader(c *HeaderContext)

	// ExitChallenge is called when exiting the challenge production.
	ExitChallenge(c *ChallengeContext)

	// ExitAuth_scheme is called when exiting the auth_scheme production.
	ExitAuth_scheme(c *Auth_schemeContext)

	// ExitAuth_params is called when exiting the auth_params production.
	ExitAuth_params(c *Auth_paramsContext)

	// ExitToken68 is called when exiting the token68 production.
	ExitToken68(c *Token68Context)

	// ExitAuth_param is called when exiting the auth_param production.
	ExitAuth_param(c *Auth_paramContext)

	// ExitAuth_lhs is called when exiting the auth_lhs production.
	ExitAuth_lhs(c *Auth_lhsContext)

	// ExitAuth_rhs is called when exiting the auth_rhs production.
	ExitAuth_rhs(c *Auth_rhsContext)

	// ExitRws is called when exiting the rws production.
	ExitRws(c *RwsContext)

	// ExitQuoted_string is called when exiting the quoted_string production.
	ExitQuoted_string(c *Quoted_stringContext)

	// ExitQd_text is called when exiting the qd_text production.
	ExitQd_text(c *Qd_textContext)

	// ExitQuoted_pair is called when exiting the quoted_pair production.
	ExitQuoted_pair(c *Quoted_pairContext)

	// ExitToken is called when exiting the token production.
	ExitToken(c *TokenContext)

	// ExitTchar is called when exiting the tchar production.
	ExitTchar(c *TcharContext)

	// ExitVchar is called when exiting the vchar production.
	ExitVchar(c *VcharContext)

	// ExitObs_text is called when exiting the obs_text production.
	ExitObs_text(c *Obs_textContext)
}
