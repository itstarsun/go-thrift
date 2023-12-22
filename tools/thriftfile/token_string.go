// Code generated by "stringer -type token -linecomment -trimprefix _"; DO NOT EDIT.

package thriftfile

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[_ILLEGAL-0]
	_ = x[_EOF-1]
	_ = x[_COMMENT-2]
	_ = x[literal_beg-3]
	_ = x[_IDENT-4]
	_ = x[_INT-5]
	_ = x[_FLOAT-6]
	_ = x[_STRING-7]
	_ = x[literal_end-8]
	_ = x[_ADD-9]
	_ = x[_SUB-10]
	_ = x[_MUL-11]
	_ = x[_QUO-12]
	_ = x[_REM-13]
	_ = x[_AND-14]
	_ = x[_OR-15]
	_ = x[_XOR-16]
	_ = x[_EQL-17]
	_ = x[_LSS-18]
	_ = x[_GTR-19]
	_ = x[_ASSIGN-20]
	_ = x[_NOT-21]
	_ = x[_LPAREN-22]
	_ = x[_LBRACK-23]
	_ = x[_LBRACE-24]
	_ = x[_COMMA-25]
	_ = x[_PERIOD-26]
	_ = x[_RPAREN-27]
	_ = x[_RBRACK-28]
	_ = x[_RBRACE-29]
	_ = x[_SEMICOLON-30]
	_ = x[_COLON-31]
	_ = x[keyword_beg-32]
	_ = x[_INCLUDE-33]
	_ = x[_NAMESPACE-34]
	_ = x[definition_beg-35]
	_ = x[_CONST-36]
	_ = x[_TYPEDEF-37]
	_ = x[_ENUM-38]
	_ = x[_STRUCT-39]
	_ = x[_UNION-40]
	_ = x[_EXCEPTION-41]
	_ = x[_SERVICE-42]
	_ = x[definition_end-43]
	_ = x[_REQUIRED-44]
	_ = x[_OPTIONAL-45]
	_ = x[_EXTENDS-46]
	_ = x[_ONEWAY-47]
	_ = x[_VOID-48]
	_ = x[_THROWS-49]
	_ = x[_MAP-50]
	_ = x[_SET-51]
	_ = x[_LIST-52]
	_ = x[keyword_end-53]
}

const _token_name = "ILLEGALEOFCOMMENTliteral_begIDENTINTFLOATSTRINGliteral_end+-*/%&|^==<>=!([{,.)]};:keyword_begincludenamespacedefinition_begconsttypedefenumstructunionexceptionservicedefinition_endrequiredoptionalextendsonewayvoidthrowsmapsetlistkeyword_end"

var _token_index = [...]uint8{0, 7, 10, 17, 28, 33, 36, 41, 47, 58, 59, 60, 61, 62, 63, 64, 65, 66, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 93, 100, 109, 123, 128, 135, 139, 145, 150, 159, 166, 180, 188, 196, 203, 209, 213, 219, 222, 225, 229, 240}

func (i token) String() string {
	if i >= token(len(_token_index)-1) {
		return "token(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _token_name[_token_index[i]:_token_index[i+1]]
}
