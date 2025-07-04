package gen

import (
	"fmt"
	"strings"
)

var (
	identNext   = 0
	identPrefix = "za"
)

func resetIdent(prefix string) {
	identPrefix = prefix
	identNext = 0
}

// generate a random identifier name
func randIdent() string {
	identNext++
	return fmt.Sprintf("%s%04d", identPrefix, identNext)
}

// This code defines the type declaration tree.
//
// Consider the following:
//
// type Marshaler struct {
// 	  Thing1 *float64 `msg:"thing1"`
// 	  Body   []byte   `msg:"body"`
// }
//
// A parser using this generator as a backend
// should parse the above into:
//
// var val Elem = &Ptr{
// 	name: "z",
// 	Value: &Struct{
// 		Name: "Marshaler",
// 		Fields: []StructField{
// 			{
// 				FieldTag: "thing1",
// 				FieldElem: &Ptr{
// 					name: "z.Thing1",
// 					Value: &BaseElem{
// 						name:    "*z.Thing1",
// 						Value:   Float64,
//						Convert: false,
// 					},
// 				},
// 			},
// 			{
// 				FieldTag: "body",
// 				FieldElem: &BaseElem{
// 					name:    "z.Body",
// 					Value:   Bytes,
// 					Convert: false,
// 				},
// 			},
// 		},
// 	},
// }

// Base is one of the
// base types
type Primitive uint8

// this is effectively the
// list of currently available
// ReadXxxx / WriteXxxx methods.
const (
	Invalid Primitive = iota
	Bytes
	String
	Float32
	Float64
	Complex64
	Complex128
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Byte
	Int
	Int8
	Int16
	Int32
	Int64
	Bool
	Intf       // interface{}
	Time       // time.Time
	Duration   // time.Duration
	Ext        // extension
	JsonNumber // json.Number

	IDENT // IDENT means an unrecognized identifier
)

// all of the recognized identities
// that map to primitive types
var primitives = map[string]Primitive{
	"[]byte":         Bytes,
	"string":         String,
	"float32":        Float32,
	"float64":        Float64,
	"complex64":      Complex64,
	"complex128":     Complex128,
	"uint":           Uint,
	"uint8":          Uint8,
	"uint16":         Uint16,
	"uint32":         Uint32,
	"uint64":         Uint64,
	"byte":           Byte,
	"rune":           Int32,
	"int":            Int,
	"int8":           Int8,
	"int16":          Int16,
	"int32":          Int32,
	"int64":          Int64,
	"bool":           Bool,
	"interface{}":    Intf,
	"any":            Intf,
	"time.Time":      Time,
	"time.Duration":  Duration,
	"msgp.Extension": Ext,
	"json.Number":    JsonNumber,
}

// types built into the library
// that satisfy all of the
// interfaces.
var builtins = map[string]struct{}{
	"msgp.Raw":    {},
	"msgp.Number": {},
}

// common data/methods for every Elem
type common struct {
	vname, alias string
	ptrRcv       bool
}

func (c *common) SetVarname(s string) { c.vname = s }
func (c *common) Varname() string     { return c.vname }
func (c *common) Alias(typ string)    { c.alias = typ }
func (c *common) hidden()             {}
func (c *common) AllowNil() bool      { return false }
func (c *common) SetIsAllowNil(bool)  {}
func (c *common) AlwaysPtr(set *bool) bool {
	if c != nil && set != nil {
		c.ptrRcv = *set
	}
	return c.ptrRcv
}

func IsPrintable(e Elem) bool {
	if be, ok := e.(*BaseElem); ok && !be.Printable() {
		return false
	}
	return true
}

// Elem is a go type capable of being
// serialized into MessagePack. It is
// implemented by *Ptr, *Struct, *Array,
// *Slice, *Map, and *BaseElem.
type Elem interface {
	// SetVarname sets this nodes
	// variable name and recursively
	// sets the names of all its children.
	// In general, this should only be
	// called on the parent of the tree.
	SetVarname(s string)

	// Varname returns the variable
	// name of the element.
	Varname() string

	// TypeName is the canonical
	// go type name of the node
	// e.g. "string", "int", "map[string]float64"
	// OR the alias name, if it has been set.
	TypeName() string

	// Alias sets a type (alias) name
	Alias(typ string)

	// Copy should perform a deep copy of the object
	Copy() Elem

	// Complexity returns a measure of the
	// complexity of element (greater than
	// or equal to 1.)
	Complexity() int

	// ZeroExpr returns the expression for the correct zero/empty
	// value.  Can be used for assignment.
	// Returns "" if zero/empty not supported for this Elem.
	ZeroExpr() string

	// AllowNil will return true for types that can be nil but doesn't automatically check.
	// This is true for slices and maps.
	AllowNil() bool

	// SetIsAllowNil will set the allownil value, if the type supports it.
	SetIsAllowNil(bool)

	// AlwaysPtr will return true if receiver should always be a pointer.
	AlwaysPtr(set *bool) bool

	// IfZeroExpr returns the expression to compare to an empty value
	// for this type, per the rules of the `omitempty` feature.
	// It is meant to be used in an if statement
	// and may include the simple statement form followed by
	// semicolon and then the expression.
	// Returns "" if zero/empty not supported for this Elem.
	// Note that this is NOT used by the `omitzero` feature.
	IfZeroExpr() string

	hidden()
}

// Ident returns the *BaseElem that corresponds
// to the provided identity.
func Ident(id string) *BaseElem {
	p, ok := primitives[id]
	if ok {
		return &BaseElem{Value: p}
	}
	be := &BaseElem{Value: IDENT}
	be.Alias(id)
	return be
}

type Array struct {
	common
	Index string // index variable name
	Size  string // array size
	Els   Elem   // child
}

func (a *Array) SetVarname(s string) {
	a.common.SetVarname(s)
ridx:
	a.Index = randIdent()

	// try to avoid using the same
	// index as a parent slice
	if strings.Contains(a.Varname(), a.Index) {
		goto ridx
	}

	a.Els.SetVarname(fmt.Sprintf("%s[%s]", a.Varname(), a.Index))
}

func (a *Array) TypeName() string {
	if a.alias != "" {
		return a.alias
	}
	a.Alias(fmt.Sprintf("[%s]%s", a.Size, a.Els.TypeName()))
	return a.alias
}

func (a *Array) Copy() Elem {
	b := *a
	b.Els = a.Els.Copy()
	return &b
}

func (a *Array) Complexity() int {
	// We consider the complexity constant and leave the children to decide on their own.
	return 2
}

// ZeroExpr returns the zero/empty expression or empty string if not supported.  Unsupported for this case.
func (a *Array) ZeroExpr() string { return "" }

// IfZeroExpr unsupported
func (a *Array) IfZeroExpr() string { return "" }

// Map is a map[string]Elem
type Map struct {
	common
	Keyidx     string // key variable name
	Validx     string // value variable name
	Value      Elem   // value element
	isAllowNil bool
}

func (m *Map) SetVarname(s string) {
	m.common.SetVarname(s)
ridx:
	m.Keyidx = randIdent()
	m.Validx = randIdent()

	// just in case
	if m.Keyidx == m.Validx {
		goto ridx
	}

	m.Value.SetVarname(m.Validx)
}

func (m *Map) TypeName() string {
	if m.alias != "" {
		return m.alias
	}
	m.Alias("map[string]" + m.Value.TypeName())
	return m.alias
}

func (m *Map) Copy() Elem {
	g := *m
	g.Value = m.Value.Copy()
	return &g
}

func (m *Map) Complexity() int {
	// Complexity of maps are considered constant. Children should decide on their own.
	return 3
}

// ZeroExpr returns the zero/empty expression or empty string if not supported.  Always "nil" for this case.
func (m *Map) ZeroExpr() string { return "nil" }

// IfZeroExpr returns the expression to compare to zero/empty.
func (m *Map) IfZeroExpr() string { return m.Varname() + " == nil" }

// AllowNil is true for maps.
func (m *Map) AllowNil() bool { return true }

// SetIsAllowNil sets whether the map is allowed to be nil.
func (m *Map) SetIsAllowNil(b bool) { m.isAllowNil = b }

type Slice struct {
	common
	Index      string
	isAllowNil bool
	Els        Elem // The type of each element
}

func (s *Slice) SetVarname(a string) {
	s.common.SetVarname(a)
	s.Index = randIdent()
	varName := s.Varname()
	if varName[0] == '*' {
		// Pointer-to-slice requires parenthesis for slicing.
		varName = "(" + varName + ")"
	}
	s.Els.SetVarname(fmt.Sprintf("%s[%s]", varName, s.Index))
}

func (s *Slice) TypeName() string {
	if s.alias != "" {
		return s.alias
	}
	s.Alias("[]" + s.Els.TypeName())
	return s.alias
}

func (s *Slice) Copy() Elem {
	z := *s
	z.Els = s.Els.Copy()
	return &z
}

func (s *Slice) Complexity() int {
	// We leave the inlining decision to the slice children.
	return 2
}

// ZeroExpr returns the zero/empty expression or empty string if not supported.  Always "nil" for this case.
func (s *Slice) ZeroExpr() string { return "nil" }

// IfZeroExpr returns the expression to compare to zero/empty.
func (s *Slice) IfZeroExpr() string { return s.Varname() + " == nil" }

// AllowNil is true for slices.
func (s *Slice) AllowNil() bool { return true }

// SetIsAllowNil sets whether the slice is allowed to be nil.
func (s *Slice) SetIsAllowNil(b bool) { s.isAllowNil = b }

// SetIsAllowNil will set whether the element is allowed to be nil.
func SetIsAllowNil(e Elem, b bool) {
	type i interface {
		SetIsAllowNil(b bool)
	}
	if x, ok := e.(i); ok {
		x.SetIsAllowNil(b)
	}
}

type Ptr struct {
	common
	Value Elem
}

func (s *Ptr) SetVarname(a string) {
	s.common.SetVarname(a)

	// struct fields are dereferenced
	// automatically...
	switch x := s.Value.(type) {
	case *Struct:
		// struct fields are automatically dereferenced
		x.SetVarname(a)
		return

	case *BaseElem:
		// identities have pointer receivers
		if x.Value == IDENT {
			// replace directive sets Convert=true and Needsref=true
			// since BaseElem is behind a pointer we set Needsref=false
			if x.Convert {
				x.Needsref(false)
			}
			x.SetVarname(a)
		} else {
			x.SetVarname("*" + a)
		}
		return

	default:
		s.Value.SetVarname("*" + a)
		return
	}
}

func (s *Ptr) TypeName() string {
	if s.alias != "" {
		return s.alias
	}
	s.Alias("*" + s.Value.TypeName())
	return s.alias
}

func (s *Ptr) Copy() Elem {
	v := *s
	v.Value = s.Value.Copy()
	return &v
}

func (s *Ptr) Complexity() int { return 1 + s.Value.Complexity() }

func (s *Ptr) Needsinit() bool {
	if be, ok := s.Value.(*BaseElem); ok && be.needsref {
		return false
	}
	return true
}

// ZeroExpr returns the zero/empty expression or empty string if not supported.  Always "nil" for this case.
func (s *Ptr) ZeroExpr() string { return "nil" }

// IfZeroExpr returns the expression to compare to zero/empty.
func (s *Ptr) IfZeroExpr() string { return s.Varname() + " == nil" }

type Struct struct {
	common
	Fields  []StructField // field list
	AsTuple bool          // write as an array instead of a map
}

func (s *Struct) TypeName() string {
	if s.alias != "" {
		return s.alias
	}
	str := "struct{\n"
	for i := range s.Fields {
		str += s.Fields[i].FieldName +
			" " + s.Fields[i].FieldElem.TypeName() +
			" " + s.Fields[i].RawTag + ";\n"
	}
	str += "}"
	s.Alias(str)
	return s.alias
}

func (s *Struct) SetVarname(a string) {
	s.common.SetVarname(a)
	writeStructFields(s.Fields, a)
}

func (s *Struct) Copy() Elem {
	g := *s
	g.Fields = make([]StructField, len(s.Fields))
	copy(g.Fields, s.Fields)
	for i := range s.Fields {
		g.Fields[i].FieldElem = s.Fields[i].FieldElem.Copy()
	}
	return &g
}

func (s *Struct) Complexity() int {
	c := 1
	for i := range s.Fields {
		c += s.Fields[i].FieldElem.Complexity()
	}
	return c
}

// ZeroExpr returns the zero/empty expression or empty string if not supported.
func (s *Struct) ZeroExpr() string {
	if s.alias == "" {
		return "" // structs with no names not supported (for now)
	}
	return "(" + s.TypeName() + "{})"
}

// IfZeroExpr returns the expression to compare to zero/empty.
func (s *Struct) IfZeroExpr() string {
	if s.alias == "" {
		return "" // structs with no names not supported (for now)
	}
	return s.Varname() + " == " + s.ZeroExpr()
}

// AnyHasTagPart returns true if HasTagPart(p) is true for any field.
func (s *Struct) AnyHasTagPart(pname string) bool {
	for _, sf := range s.Fields {
		if sf.HasTagPart(pname) {
			return true
		}
	}
	return false
}

// CountFieldTagPart the count of HasTagPart(p) is true for any field.
func (s *Struct) CountFieldTagPart(pname string) int {
	var n int
	for _, sf := range s.Fields {
		if sf.HasTagPart(pname) {
			n++
		}
	}
	return n
}

type StructField struct {
	FieldTag      string   // the string inside the `msg:""` tag up to the first comma
	FieldTagParts []string // the string inside the `msg:""` tag split by commas
	RawTag        string   // the full struct tag
	FieldName     string   // the name of the struct field
	FieldElem     Elem     // the field type
}

// HasTagPart returns true if the specified tag part (option) is present.
func (sf *StructField) HasTagPart(pname string) bool {
	if len(sf.FieldTagParts) < 2 {
		return false
	}
	for _, p := range sf.FieldTagParts[1:] {
		if p == pname {
			return true
		}
	}
	return false
}

type ShimMode int

const (
	Cast ShimMode = iota
	Convert
)

// BaseElem is an element that
// can be represented by a primitive
// MessagePack type.
type BaseElem struct {
	common
	ShimMode     ShimMode  // Method used to shim
	ShimToBase   string    // shim to base type, or empty
	ShimFromBase string    // shim from base type, or empty
	Value        Primitive // Type of element
	Convert      bool      // should we do an explicit conversion?
	zerocopy     bool      // Allow zerocopy for byte slices in unmarshal.
	mustinline   bool      // must inline; not printable
	needsref     bool      // needs reference for shim
	allowNil     *bool     // Override from parent.
}

func (s *BaseElem) Printable() bool { return !s.mustinline }

func (s *BaseElem) Alias(typ string) {
	s.common.Alias(typ)
	if s.Value != IDENT {
		s.Convert = true
	}
	if strings.Contains(typ, ".") {
		s.mustinline = true
	}
}

func (s *BaseElem) AllowNil() bool {
	if s.allowNil == nil {
		return s.Value == Bytes
	}
	return *s.allowNil
}

// SetIsAllowNil will override allownil when tag has been parsed.
func (s *BaseElem) SetIsAllowNil(b bool) {
	s.allowNil = &b
}

func (s *BaseElem) SetVarname(a string) {
	// extensions whose parents
	// are not pointers need to
	// be explicitly referenced
	if s.Value == Ext || s.needsref {
		if strings.HasPrefix(a, "*") {
			s.common.SetVarname(a[1:])
			return
		}
		s.common.SetVarname("&" + a)
		return
	}

	s.common.SetVarname(a)
}

// TypeName returns the syntactically correct Go
// type name for the base element.
func (s *BaseElem) TypeName() string {
	if s.alias != "" {
		return s.alias
	}
	s.common.Alias(s.BaseType())
	return s.alias
}

// ToBase, used if Convert==true, is used as tmp = {{ToBase}}({{Varname}})
func (s *BaseElem) ToBase() string {
	if s.ShimToBase != "" {
		return s.ShimToBase
	}
	return s.BaseType()
}

// FromBase, used if Convert==true, is used as {{Varname}} = {{FromBase}}(tmp)
func (s *BaseElem) FromBase() string {
	if s.ShimFromBase != "" {
		return s.ShimFromBase
	}
	return s.TypeName()
}

// BaseName returns the string form of the
// base type (e.g. Float64, Ident, etc)
func (s *BaseElem) BaseName() string {
	// time.Time and time.Duration are special cases;
	// we strip the package prefix
	if s.Value == Time {
		return "Time"
	}
	if s.Value == Duration {
		return "Duration"
	}
	if s.Value == JsonNumber {
		return "JSONNumber"
	}
	return s.Value.String()
}

func (s *BaseElem) BaseType() string {
	switch s.Value {
	case IDENT:
		return s.TypeName()

	// exceptions to the naming/capitalization
	// rule:
	case Intf:
		return "interface{}"
	case Bytes:
		return "[]byte"
	case Time:
		return "time.Time"
	case Duration:
		return "time.Duration"
	case JsonNumber:
		return "json.Number"
	case Ext:
		return "msgp.Extension"

	// everything else is base.String() with
	// the first letter as lowercase
	default:
		return strings.ToLower(s.BaseName())
	}
}

func (s *BaseElem) Needsref(b bool) {
	s.needsref = b
}

func (s *BaseElem) Copy() Elem {
	g := *s
	return &g
}

func (s *BaseElem) Complexity() int {
	if s.Convert && !s.mustinline {
		return 2
	}
	// we need to return 1 if !printable(),
	// in order to make sure that stuff gets
	// inlined appropriately
	return 1
}

// Resolved returns whether or not
// the type of the element is
// a primitive or a builtin provided
// by the package.
func (s *BaseElem) Resolved() bool {
	if s.Value == IDENT {
		_, ok := builtins[s.TypeName()]
		return ok
	}
	return true
}

// ZeroExpr returns the zero/empty expression or empty string if not supported.
func (s *BaseElem) ZeroExpr() string {
	switch s.Value {
	case Bytes:
		return "nil"
	case String:
		return "\"\""
	case Complex64, Complex128:
		return "complex(0,0)"
	case Float32,
		Float64,
		Uint,
		Uint8,
		Uint16,
		Uint32,
		Uint64,
		Byte,
		Int,
		Int8,
		Int16,
		Int32,
		Int64,
		Duration:
		return "0"
	case Bool:
		return "false"
	case Time:
		return "(time.Time{})"
	case JsonNumber:
		return `""`
	case Intf:
		return "nil"
	}

	return ""
}

// IfZeroExpr returns the expression to compare to zero/empty.
func (s *BaseElem) IfZeroExpr() string {
	z := s.ZeroExpr()
	if z == "" {
		return ""
	}
	return s.Varname() + " == " + z
}

func (k Primitive) String() string {
	switch k {
	case String:
		return "String"
	case Bytes:
		return "Bytes"
	case Float32:
		return "Float32"
	case Float64:
		return "Float64"
	case Complex64:
		return "Complex64"
	case Complex128:
		return "Complex128"
	case Uint:
		return "Uint"
	case Uint8:
		return "Uint8"
	case Uint16:
		return "Uint16"
	case Uint32:
		return "Uint32"
	case Uint64:
		return "Uint64"
	case Byte:
		return "Byte"
	case Int:
		return "Int"
	case Int8:
		return "Int8"
	case Int16:
		return "Int16"
	case Int32:
		return "Int32"
	case Int64:
		return "Int64"
	case Bool:
		return "Bool"
	case Intf:
		return "Intf"
	case Time:
		return "time.Time"
	case Duration:
		return "time.Duration"
	case Ext:
		return "Extension"
	case JsonNumber:
		return "json.Number"
	case IDENT:
		return "Ident"
	default:
		return "INVALID"
	}
}

// writeStructFields is a trampoline for writeBase for
// all of the fields in a struct
func writeStructFields(s []StructField, name string) {
	for i := range s {
		s[i].FieldElem.SetVarname(fmt.Sprintf("%s.%s", name, s[i].FieldName))
	}
}

// coerceArraySize ensures we can compare constant array lengths.
//
// msgpack array headers are 32 bit unsigned, which is reflected in the
// ArrayHeader implementation in this library using uint32. On the Go side, we
// can declare array lengths as any constant integer width, which breaks when
// attempting a direct comparison to an array header's uint32.
func coerceArraySize(asz string) string {
	return fmt.Sprintf("uint32(%s)", asz)
}
