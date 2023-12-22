package thriftfile

import (
	"fmt"
	gotoken "go/token"
	"sort"
)

// Parse parses a .thrift file.
func Parse(fset *gotoken.FileSet, filename string, src []byte) (f *File, err error) {
	var p parser
	defer func() {
		p.errs.sort()
		err = p.errs.err()
	}()
	p.init(fset, filename, src)
	f = p.file_()
	return f, err
}

type error_ struct {
	pos gotoken.Position
	msg string
}

func (e error_) Error() string {
	return e.pos.String() + ": " + e.msg
}

type errors_ []*error_

func (p *errors_) add(pos gotoken.Position, msg string) {
	*p = append(*p, &error_{pos, msg})
}

func (p errors_) err() error {
	if len(p) == 0 {
		return nil
	}
	return p
}

func (p errors_) sort() {
	sort.Sort(p)
}

func (p errors_) Len() int      { return len(p) }
func (p errors_) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p errors_) Less(i, j int) bool {
	e := &p[i].pos
	f := &p[j].pos
	// Note that it is not sufficient to simply compare file offsets because
	// the offsets do not reflect modified line information (through //line
	// comments).
	if e.Filename != f.Filename {
		return e.Filename < f.Filename
	}
	if e.Line != f.Line {
		return e.Line < f.Line
	}
	if e.Column != f.Column {
		return e.Column < f.Column
	}
	return p[i].msg < p[j].msg
}

func (p errors_) Error() string {
	switch len(p) {
	case 0:
		return "no errors"
	case 1:
		return p[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", p[0], len(p)-1)
}

type parser struct {
	scanner
	errs errors_
}

func (p *parser) init(fset *gotoken.FileSet, filename string, src []byte) {
	p.scanner.init(fset.AddFile(filename, -1, len(src)), src, func(pos gotoken.Position, msg string) {
		p.error(p.file.Pos(pos.Offset), msg)
	})
	p.next()
}

func (p *parser) error(pos gotoken.Pos, msg string) {
	p.errs.add(p.file.Position(pos), msg)
}

func (p *parser) errorExpected(pos gotoken.Pos, msg string) {
	msg = "expected " + msg
	if pos == p.pos {
		switch {
		case p.tok.isLiteral():
			msg += ", found " + p.scanner.lit()
		default:
			msg += ", found '" + p.tok.String() + "'"
		}
	}
	p.error(pos, msg)
}

func (p *parser) got(tok token) bool {
	if p.tok == tok {
		p.next()
		return true
	}
	return false
}

func (p *parser) want(tok token) {
	if !p.got(tok) {
		p.errorExpected(p.pos, "'"+tok.String()+"'")
		p.next()
	}
}

func (p *parser) next0() {
	p.scanner.next()
}

func (p *parser) next() {
	p.next0()
	for p.tok == _COMMENT {
		p.next0()
	}
}

func (p *parser) gotAssign() bool {
	if p.got(_ASSIGN) {
		return true
	}
	if p.tok == _EQL {
		p.want(_ASSIGN)
		return true
	}
	return false
}

func (p *parser) list(open, close token, f func()) {
	p.want(open)
	for p.tok != _EOF && p.tok != close {
		f()
		p.separator()
	}
	p.want(close)
}

func (p *parser) lit(tok token) string {
	var lit string
	if p.tok == tok {
		lit = p.scanner.lit()
		p.next()
	} else {
		p.want(tok)
	}
	return lit
}

func (p *parser) ident() *Ident {
	var x Ident
	x.Name = p.lit(_IDENT)
	return &x
}

func (p *parser) int() *Int {
	var x Int
	x.Value = p.lit(_INT)
	return &x
}

func (p *parser) float() *Float {
	var x Float
	x.Value = p.lit(_FLOAT)
	return &x
}

func (p *parser) string() *String {
	var x String
	x.Value = p.lit(_STRING)
	return &x
}

func (p *parser) file_() *File {
	var x File
	for h := p.header(); h != nil; h = p.header() {
		x.Headers = append(x.Headers, h)
	}
	for p.tok != _EOF {
		x.Defs = append(x.Defs, p.def())
	}
	return &x
}

func (p *parser) header() Header {
	switch p.tok {
	case _INCLUDE:
		return p.include()
	case _NAMESPACE:
		return p.namespace()
	default:
		return nil
	}
}

func (p *parser) include() *Include {
	var x Include
	p.want(_INCLUDE)
	x.Path = p.string()
	return &x
}

func (p *parser) namespace() *Namespace {
	var x Namespace
	p.want(_NAMESPACE)
	if !p.got(_MUL) {
		x.Scope = p.ident()
	}
	x.Name = p.ident()
	x.Annotations = p.annotations()
	return &x
}

type badDef struct{}

func (p *parser) def() (d Def) {
	pos := p.pos
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(badDef); !ok {
				panic(r)
			}
			for p.tok != _EOF && !p.tok.isDefinition() {
				p.next()
			}
			d = &BadDef{
				From: pos,
				To:   p.pos,
			}
		}
	}()

	switch p.tok {
	case _CONST:
		return p.const_()
	case _TYPEDEF:
		return p.typedef()
	case _ENUM:
		return p.enum()
	case _STRUCT, _UNION, _EXCEPTION:
		return p.struct_(p.tok)
	case _SERVICE:
		return p.service()
	default:
		p.errorExpected(p.pos, "definition")
		panic(badDef{})
	}
}

func (p *parser) const_() *Const {
	var x Const
	p.want(_CONST)
	x.Type = p.type_()
	x.Name = p.ident()
	p.want(_ASSIGN)
	x.Value = p.value()
	p.separator()
	return &x
}

func (p *parser) typedef() *Typedef {
	var x Typedef
	p.want(_TYPEDEF)
	x.Type = p.type_()
	x.Name = p.ident()
	x.Annotations = p.annotations()
	p.separator()
	return &x
}

func (p *parser) enum() *Enum {
	var x Enum
	p.want(_ENUM)
	x.Name = p.ident()
	x.Values = p.enumValues()
	x.Annotations = p.annotations()
	return &x
}

func (p *parser) enumValue() *EnumValue {
	var x EnumValue
	x.Name = p.ident()
	if p.gotAssign() {
		x.Value = p.int()
	}
	x.Annotations = p.annotations()
	return &x
}

func (p *parser) enumValues() *EnumValueList {
	var x EnumValueList
	p.list(_LBRACE, _RBRACE, func() {
		x.List = append(x.List, p.enumValue())
	})
	return &x
}

func (p *parser) struct_(tok token) *Struct {
	var x Struct
	p.want(tok)
	x.Union = tok == _UNION
	x.Exception = tok == _EXCEPTION
	x.Name = p.ident()
	x.Fields = p.fields(_LBRACE, _RBRACE)
	x.Annotations = p.annotations()
	return &x
}

func (p *parser) fieldName() *Ident {
	var x Ident
	if p.tok == _IDENT || p.tok.isKeyword() {
		x.Name = p.scanner.lit()
		p.next()
	} else {
		p.want(_IDENT)
	}
	return &x
}

func (p *parser) field() *Field {
	var x Field
	if p.tok == _INT {
		x.ID = p.int()
		p.want(_COLON)
	}
	switch {
	case p.got(_REQUIRED):
		x.Required = true
	case p.got(_OPTIONAL):
		x.Optional = true
	}
	x.Type = p.type_()
	x.Reference = p.got(_AND)
	x.Name = p.fieldName()
	if p.gotAssign() {
		x.Default = p.value()
	}
	x.Annotations = p.annotations()
	return &x
}

func (p *parser) fields(open, close token) *FieldList {
	var x FieldList
	p.list(open, close, func() {
		x.List = append(x.List, p.field())
	})
	return &x
}

func (p *parser) service() *Service {
	var x Service
	p.want(_SERVICE)
	x.Name = p.ident()
	if p.got(_EXTENDS) {
		x.Extends = p.ident()
	}
	x.Methods = p.methods()
	x.Annotations = p.annotations()
	return &x
}

func (p *parser) method() *Method {
	var x Method
	if p.got(_ONEWAY) {
		x.OneWay = true
	}
	if p.got(_VOID) {
		x.Void = true
	} else {
		x.Return = p.type_()
	}
	x.Name = p.ident()
	x.Arguments = p.fields(_LPAREN, _RPAREN)
	if p.got(_THROWS) {
		x.Throws = p.fields(_LPAREN, _RPAREN)
	}
	x.Annotations = p.annotations()
	return &x
}

func (p *parser) methods() *MethodList {
	var x MethodList
	p.list(_LBRACE, _RBRACE, func() {
		x.List = append(x.List, p.method())
	})
	return &x
}

func (p *parser) type_() Type {
	switch p.tok {
	case _MAP:
		return p.map_()
	case _SET:
		return p.set()
	case _LIST:
		return p.list_()
	case _IDENT:
		return p.typeRef()
	default:
		p.errorExpected(p.pos, "type")
		panic(badDef{})
	}
}

func (p *parser) typeRef() *TypeRef {
	var x TypeRef
	x.Name = p.ident()
	x.Annotations = p.annotations()
	return &x
}

func (p *parser) map_() *Map {
	var x Map
	p.want(_MAP)
	p.want(_LSS)
	x.Key = p.type_()
	p.want(_COMMA)
	x.Value = p.type_()
	p.want(_GTR)
	x.Annotations = p.annotations()
	return &x
}

func (p *parser) set() *Set {
	var x Set
	p.want(_SET)
	p.want(_LSS)
	x.Element = p.type_()
	p.want(_GTR)
	x.Annotations = p.annotations()
	return &x
}

func (p *parser) list_() *List {
	var x List
	p.want(_LIST)
	p.want(_LSS)
	x.Element = p.type_()
	p.want(_GTR)
	x.Annotations = p.annotations()
	return &x
}

func (p *parser) value() Value {
	switch p.tok {
	case _IDENT:
		return p.ident()
	case _INT:
		return p.int()
	case _FLOAT:
		return p.float()
	case _STRING:
		return p.string()
	case _LBRACK:
		return p.values()
	case _LBRACE:
		return p.keyValuePairs()
	default:
		p.errorExpected(p.pos, "value")
		panic(badDef{})
	}
}

func (p *parser) values() *ValueList {
	var x ValueList
	p.list(_LBRACK, _RBRACK, func() {
		x.List = append(x.List, p.value())
	})
	return &x
}

func (p *parser) keyValuePair() *KeyValuePair {
	var x KeyValuePair
	x.Key = p.value()
	p.want(_COLON)
	x.Value = p.value()
	return &x
}

func (p *parser) keyValuePairs() *KeyValuePairList {
	var x KeyValuePairList
	p.list(_LBRACE, _RBRACE, func() {
		x.List = append(x.List, p.keyValuePair())
	})
	return &x
}

func (p *parser) annotation() *Annotation {
	var x Annotation
	x.Name = p.ident()
	if p.gotAssign() {
		x.Value = p.string()
	}
	return &x
}

func (p *parser) annotations() *AnnotationList {
	if p.tok != _LPAREN {
		return nil
	}
	var x AnnotationList
	p.list(_LPAREN, _RPAREN, func() {
		x.List = append(x.List, p.annotation())
	})
	return &x
}

func (p *parser) separator() {
	for p.tok == _COMMA || p.tok == _SEMICOLON {
		p.next()
	}
}
