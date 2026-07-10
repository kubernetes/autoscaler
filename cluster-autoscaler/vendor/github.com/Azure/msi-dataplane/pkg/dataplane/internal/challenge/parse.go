package challenge

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

type Listener struct {
	challenges []Challenge
	errors     []error
	*BaseChallengeListener
}

// EnterChallenge is called when production challenge is entered.
func (s *Listener) EnterChallenge(ctx *ChallengeContext) {
	challenge := Challenge{
		Scheme:     ctx.Auth_scheme().GetText(),
		Parameters: map[string]string{},
	}
	for _, list := range ctx.AllAuth_params() {
		for _, param := range list.AllAuth_param() {
			rhs := param.Auth_rhs().GetText()
			if param.Auth_rhs().Quoted_string() != nil {
				value, err := strconv.Unquote(param.Auth_rhs().Quoted_string().GetText())
				if err != nil {
					s.errors = append(s.errors, fmt.Errorf("failed to unquote %s: %w", param.Auth_rhs().Quoted_string().GetText(), err))
					return
				}
				rhs = value
			}
			challenge.Parameters[param.Auth_lhs().GetText()] = rhs
		}
		for _, value := range ctx.AllToken68() {
			challenge.Values = append(challenge.Values, value.GetText())
		}
	}
	s.challenges = append(s.challenges, challenge)
}

type errorSink struct {
	*antlr.DefaultErrorListener
	errors []error
}

func (e *errorSink) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, exception antlr.RecognitionException) {
	e.errors = append(e.errors, fmt.Errorf("syntax error in line %d:%d: %v: %s", line, column, offendingSymbol, msg))
}

var _ antlr.ErrorListener = (*errorSink)(nil)

func Parse(headers http.Header) ([]Challenge, error) {
	var challenges []Challenge
	var errors []error
	parsingErrors := &errorSink{}
	for _, value := range headers.Values("WWW-Authenticate") {
		p := NewChallengeParser(
			antlr.NewCommonTokenStream(
				NewChallengeLexer(
					antlr.NewInputStream(value),
				),
				0,
			),
		)
		p.AddErrorListener(parsingErrors)
		listener := &Listener{}
		antlr.ParseTreeWalkerDefault.Walk(listener, p.Header())
		challenges = append(challenges, listener.challenges...)
		errors = append(errors, listener.errors...)
	}
	if parsingErrors.errors != nil {
		return nil, collapseErrors(parsingErrors.errors)
	}
	if errors != nil {
		return nil, collapseErrors(errors)
	}
	return challenges, nil
}

func collapseErrors(errors []error) error {
	var reasons []string
	for _, err := range errors {
		reasons = append(reasons, err.Error())
	}
	return fmt.Errorf("parsing failed: %s", strings.Join(reasons, ","))
}

type Challenge struct {
	Scheme     string            `json:"scheme"`
	Parameters map[string]string `json:"parameters"`
	Values     []string          `json:"values"`
}
