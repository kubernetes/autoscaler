// $antlr-format alignTrailingComments true, columnLimit 150, minEmptyLines 1, maxEmptyLinesToKeep 1, reflowComments false, useTab false
// $antlr-format allowShortRulesOnASingleLine false, allowShortBlocksOnASingleLine true, alignSemicolons hanging, alignColons hanging

/*
This grammar is based on the HTTP/1.1 specification (RFC 9110).
*/
grammar Challenge;

/*
WWW-Authenticate = #challenge
ref: https://www.rfc-editor.org/rfc/rfc9110.html#section-11.6.1

# in #auth-param is a list
1#element => element *( OWS "," OWS element )
#element => [ 1#element ]
ref: https://www.rfc-editor.org/rfc/rfc9110.html#section-5.6.1

   Square brackets enclose an optional element sequence:
         [foo bar]
   is equivalent to
         *1(foo bar).
https://www.rfc-editor.org/rfc/rfc5234.html#section-3.8
*/
header: challenge ((SP | HTAB)* COMMA (SP | HTAB)* challenge)*;

/*
challenge = auth-scheme [ 1*SP ( token68 / #auth-param ) ]
ref: https://www.rfc-editor.org/rfc/rfc9110.html#section-11.3
*/
challenge: auth_scheme (SP (token68 | auth_params?))*;

/*
auth-scheme = token
ref: https://www.rfc-editor.org/rfc/rfc9110.html#section-11.2
*/
auth_scheme: token;

/*
# in #auth-param is a list
1#element => element *( OWS "," OWS element )
#element => [ 1#element ]
ref: https://www.rfc-editor.org/rfc/rfc9110.html#section-5.6.1

   Square brackets enclose an optional element sequence:
         [foo bar]
   is equivalent to
         *1(foo bar).
https://www.rfc-editor.org/rfc/rfc5234.html#section-3.8
*/
auth_params: auth_param ((SP | HTAB)* COMMA (SP | HTAB)* auth_param)*;

/*
token68 = 1*( ALPHA / DIGIT / "-" / "." / "_" / "~" / "+" / "/" ) *"="
ref: https://www.rfc-editor.org/rfc/rfc9110.html#section-11.2
*/
token68: (ALPHA | DIGIT | MINUS | COMMA | UNDERSCORE | TILDE | PLUS | SLASH)+ EQUALS*;

/*
auth-param = token BWS "=" BWS ( token / quoted-string )
ref: https://www.rfc-editor.org/rfc/rfc9110.html#section-11.2
*/
auth_param: auth_lhs (SP | HTAB)* EQUALS (SP | HTAB)* (auth_rhs);

auth_lhs: token;
auth_rhs: token | quoted_string;

/*
RWS = 1*( SP / HTAB ) ; required whitespace
https://www.rfc-editor.org/rfc/rfc9110.html#section-5.6.3
*/
rws: (SP | HTAB)+;

/*
quoted-string = DQUOTE *( qdtext / quoted-pair ) DQUOTE
ref: https://www.rfc-editor.org/rfc/rfc9110.html#section-5.6.4
*/
quoted_string: DQUOTE (qd_text | quoted_pair)+ DQUOTE;

/*
qdtext = HTAB / SP / %x21 / %x23-5B / %x5D-7E / obs-text
ref: https://www.rfc-editor.org/rfc/rfc9110.html#section-5.6.4
*/
qd_text
    : HTAB
    | SP
    | EXCLAMATION_MARK
    | HASH
    | DOLLAR
    | PERCENT
    | AMPERSAND
    | SQUOTE
    | OPEN_PARENS
    | CLOSE_PARENS
    | ASTERISK
    | PLUS
    | COMMA
    | MINUS
    | PERIOD
    | SLASH
    | DIGIT
    | COLON
    | SEMICOLON
    | LESS_THAN
    | EQUALS
    | GREATER_THAN
    | QUESTION
    | AT
    | OPEN_BRACKET
    | CLOSE_BRACKET
    | CARET
    | UNDERSCORE
    | GRAVE
    | ALPHA
    | OPEN_BRACE
    | PIPE
    | CLOSE_BRACE
    | TILDE
    | obs_text;

/*
quoted-pair = "\" ( HTAB / SP / VCHAR / obs-text )
ref: https://www.rfc-editor.org/rfc/rfc9110.html#section-5.6.4
*/
quoted_pair: BACKSLASH (HTAB | SP | vchar | obs_text);

/*
token = 1*tchar
*/
token: tchar+;

/*
tchar = "!" / "#" / "$" / "%" / "&" / "'" / "*" / "+" / "-" / "." / "^" / "_" / "`" / "|" / "~" / DIGIT / ALPHA
any VCHAR, except delimiters
ref: https://www.rfc-editor.org/rfc/rfc9110.html#section-5.6.2
*/
tchar
    : EXCLAMATION_MARK
    | HASH
    | DOLLAR
    | PERCENT
    | AMPERSAND
    | SQUOTE
    | ASTERISK
    | PLUS
    | MINUS
    | PERIOD
    | CARET
    | UNDERSCORE
    | GRAVE
    | PIPE
    | TILDE
    | DIGIT
    | ALPHA
    ;

/*
 VCHAR = %x21-7E ; visible (printing) characters
 */
vchar: EXCLAMATION_MARK |
       DQUOTE |
       HASH |
       DOLLAR |
       PERCENT |
       AMPERSAND |
       SQUOTE |
       OPEN_PARENS |
       CLOSE_PARENS |
       ASTERISK |
       PLUS |
       COMMA |
       MINUS |
       PERIOD |
       SLASH |
       DIGIT |
       COLON |
       SEMICOLON |
       LESS_THAN |
       EQUALS |
       GREATER_THAN |
       QUESTION |
       AT |
       OPEN_BRACKET |
       BACKSLASH |
       CLOSE_BRACKET |
       CARET |
       UNDERSCORE |
       GRAVE |
       ALPHA |
       OPEN_BRACE |
       PIPE |
       CLOSE_BRACE |
       TILDE |
       EXTENDED_ASCII;

/*
 OBS_TEXT = %x80-FF
 ref: https://www.rfc-editor.org/rfc/rfc9110.html#section-5.5
*/
obs_text: EXTENDED_ASCII;

/*
ASCII primitives:
*/
HTAB: '\t';
SP: ' ';
EXCLAMATION_MARK: '!';
DQUOTE: '"';
HASH: '#';
DOLLAR: '$';
PERCENT: '%';
AMPERSAND: '&';
SQUOTE: '\'';
OPEN_PARENS: '(';
CLOSE_PARENS: ')';
ASTERISK: '*';
PLUS: '+';
COMMA: ',';
MINUS: '-';
PERIOD: '.';
SLASH: '/';
DIGIT: [0-9];
COLON: ':';
SEMICOLON: ';';
LESS_THAN: '<';
EQUALS: '=';
GREATER_THAN: '>';
QUESTION: '?';
AT: '@';
ALPHA: [A-Za-z];
OPEN_BRACKET: '[';
BACKSLASH: '\\';
CLOSE_BRACKET: ']';
CARET: '^';
UNDERSCORE: '_';
GRAVE: '`';
OPEN_BRACE: '{';
PIPE: '|';
CLOSE_BRACE: '}';
TILDE: '~';
EXTENDED_ASCII: '\u0080' .. '\u00ff';