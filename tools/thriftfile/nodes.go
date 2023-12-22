package thriftfile

import (
	gotoken "go/token"
)

// A Node represents a syntax tree node.
type Node interface {
	aNode()
}

type node struct{}

func (*node) aNode() {}

// A File node represents a .thrift file.
type File struct {
	Headers []Header
	Defs    []Def
	node
}

// A Header represents a header node.
type Header interface {
	Node
	aHeader()
}

type headerNode struct {
	node
}

func (*headerNode) aHeader() {}

// An Include node represents an include header.
type Include struct {
	Path *String
	headerNode
}

// A Namespace node represents a namespace header.
type Namespace struct {
	Scope       *Ident // nil means all scopes
	Name        *Ident
	Annotations *AnnotationList // may be nil
	headerNode
}

// A Def represents a definition node.
type Def interface {
	Node
	aDef()
}

type defNode struct {
	node
}

func (*defNode) aDef() {}

// A BadDef node is a placeholder for a definition containing
// syntax errors for which a correct definition node cannot be
// created.
type BadDef struct {
	From, To gotoken.Pos
	defNode
}

// A Const node represents a constant definition.
type Const struct {
	Type  Type
	Name  *Ident
	Value Value
	defNode
}

// A Typedef node represents a type definition.
type Typedef struct {
	Type        Type
	Name        *Ident
	Annotations *AnnotationList // may be nil
	defNode
}

// An Enum node represents an enum definition.
type Enum struct {
	Name        *Ident
	Values      *EnumValueList
	Annotations *AnnotationList // may be nil
	defNode
}

// An EnumValue node represents an enum value definition.
type EnumValue struct {
	Name        *Ident
	Value       *Int            // may be nil
	Annotations *AnnotationList // may be nil
	node
}

// An EnumValueList node represents a list of enum value definitions
// enclosed by curly braces.
type EnumValueList struct {
	List []*EnumValue
	node
}

// A Struct node represents a struct-like definition.
type Struct struct {
	Union       bool
	Exception   bool
	Name        *Ident
	Fields      *FieldList
	Annotations *AnnotationList // may be nil
	defNode
}

// A Field node represents a field definition.
type Field struct {
	ID          *Int // may be nil
	Required    bool
	Optional    bool
	Type        Type
	Reference   bool
	Name        *Ident
	Default     Value           // may be nil
	Annotations *AnnotationList // may be nil
	node
}

// A FieldList node represents a list of field definitions
// enclosed by curly braces or parentheses.
type FieldList struct {
	List []*Field
	node
}

// A Service node represents a service definition.
type Service struct {
	Name        *Ident
	Extends     *Ident // may be nil
	Methods     *MethodList
	Annotations *AnnotationList // may be nil
	defNode
}

// A Method node represents a method definition.
type Method struct {
	OneWay      bool
	Void        bool
	Return      Type // may be nil
	Name        *Ident
	Arguments   *FieldList
	Throws      *FieldList      // may be nil
	Annotations *AnnotationList // may be nil
	node
}

// A MethodList node represents a list of method definitions
// enclosed by curly braces.
type MethodList struct {
	List []*Method
	node
}

// A Type represents a type node.
type Type interface {
	Node
	aType()
}

type typeNode struct {
	node
}

func (*typeNode) aType() {}

// A Map node represents a map type.
type Map struct {
	Key         Type
	Value       Type
	Annotations *AnnotationList // may be nil
	typeNode
}

// A Set node represents a set type.
type Set struct {
	Element     Type
	Annotations *AnnotationList // may be nil
	typeNode
}

// A List node represents a list type.
type List struct {
	Element     Type
	Annotations *AnnotationList // may be nil
	typeNode
}

// A TypeRef node represents a type reference.
type TypeRef struct {
	Name        *Ident
	Annotations *AnnotationList // may be nil
	typeNode
}

// A Value represents a value node.
type Value interface {
	Node
	aValue()
}

type valueNode struct {
	node
}

func (*valueNode) aValue() {}

// A ValueList node represents a list of values
// enclosed by square brackets.
type ValueList struct {
	List []Value
	valueNode
}

// A KeyValuePair node represents a key-value pair.
type KeyValuePair struct {
	Key, Value Value
	node
}

// A KeyValuePairList node represents a list of key-value pairs
// enclosed by curly braces.
type KeyValuePairList struct {
	List []*KeyValuePair
	valueNode
}

// An Ident node represents an identifier.
type Ident struct {
	Name string
	valueNode
}

// An Int node represents an integer literal.
type Int struct {
	Value string
	valueNode
}

// A Float node represents a floating-point literal.
type Float struct {
	Value string
	valueNode
}

// A String node represents a string literal.
type String struct {
	Value string
	valueNode
}

// An Annotation represents an annotation.
type Annotation struct {
	Name  *Ident
	Value *String // may be nil
	node
}

// An AnnotationList node represents a list of annotations
// enclosed by parentheses.
type AnnotationList struct {
	List []*Annotation
	node
}
