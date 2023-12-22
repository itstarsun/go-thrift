package thriftfile

//go:generate stringer -type token -linecomment -trimprefix _
type token uint8

const (
	_ILLEGAL token = iota
	_EOF
	_COMMENT

	literal_beg

	_IDENT
	_INT
	_FLOAT
	_STRING

	literal_end

	_ADD // +
	_SUB // -
	_MUL // *
	_QUO // /
	_REM // %

	_AND // &
	_OR  // |
	_XOR // ^

	_EQL    // ==
	_LSS    // <
	_GTR    // >
	_ASSIGN // =
	_NOT    // !

	_LPAREN // (
	_LBRACK // [
	_LBRACE // {
	_COMMA  // ,
	_PERIOD // .

	_RPAREN    // )
	_RBRACK    // ]
	_RBRACE    // }
	_SEMICOLON // ;
	_COLON     // :

	keyword_beg

	_INCLUDE   // include
	_NAMESPACE // namespace

	definition_beg

	_CONST     // const
	_TYPEDEF   // typedef
	_ENUM      // enum
	_STRUCT    // struct
	_UNION     // union
	_EXCEPTION // exception
	_SERVICE   // service

	definition_end

	_REQUIRED // required
	_OPTIONAL // optional

	_EXTENDS // extends

	_ONEWAY // oneway
	_VOID   // void
	_THROWS // throws

	_MAP  // map
	_SET  // set
	_LIST // list

	keyword_end
)

var keywords map[string]token

func init() {
	keywords = make(map[string]token, keyword_end-(keyword_beg+1))
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[i.String()] = i
	}
}

func (tok token) isLiteral() bool    { return literal_beg < tok && tok < literal_end }
func (tok token) isKeyword() bool    { return keyword_beg < tok && tok < keyword_end }
func (tok token) isDefinition() bool { return definition_beg < tok && tok < definition_end }
