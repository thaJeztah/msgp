MessagePack Code Generator
=======

[![test](https://github.com/tinylib/msgp/actions/workflows/test.yml/badge.svg)](https://github.com/tinylib/msgp/actions/workflows/test.yml)
[![validate](https://github.com/tinylib/msgp/actions/workflows/validate.yml/badge.svg)](https://github.com/tinylib/msgp/actions/workflows/validate.yml)

This is a code generation tool and serialization library for [MessagePack](http://msgpack.org). You can read more about MessagePack [in the wiki](http://github.com/tinylib/msgp/wiki), or at [msgpack.org](http://msgpack.org).

### Why?

- Use Go as your schema language
- Performance
- [JSON interop](http://godoc.org/github.com/tinylib/msgp/msgp#CopyToJSON)
- [User-defined extensions](http://github.com/tinylib/msgp/wiki/Using-Extensions)
- Type safety
- Encoding flexibility

### Quickstart

In a source file, include the following directive:

```go
//go:generate msgp
```

The `msgp` command will generate serialization methods for all exported type declarations in the file.

You can [read more about the code generation options here](http://github.com/tinylib/msgp/wiki/Using-the-Code-Generator).

### Use

Field names can be set in much the same way as the `encoding/json` package. For example:

```go
type Person struct {
	Name       string `msg:"name"`
	Address    string `msg:"address"`
	Age        int    `msg:"age"`
	Hidden     string `msg:"-"` // this field is ignored
	unexported bool             // this field is also ignored
}
```

By default, the code generator will satisfy `msgp.Sizer`, `msgp.Encodable`, `msgp.Decodable`, 
`msgp.Marshaler`, and `msgp.Unmarshaler`. Carefully-designed applications can use these methods to do
marshalling/unmarshalling with zero heap allocations.

While `msgp.Marshaler` and `msgp.Unmarshaler` are quite similar to the standard library's
`json.Marshaler` and `json.Unmarshaler`, `msgp.Encodable` and `msgp.Decodable` are useful for 
stream serialization. (`*msgp.Writer` and `*msgp.Reader` are essentially protocol-aware versions
of `*bufio.Writer` and `*bufio.Reader`, respectively.)

### Features

 - Extremely fast generated code
 - Test and benchmark generation
 - JSON interoperability (see `msgp.CopyToJSON() and msgp.UnmarshalAsJSON()`)
 - Support for complex type declarations
 - Native support for Go's `time.Time`, `complex64`, and `complex128` types 
 - Generation of both `[]byte`-oriented and `io.Reader/io.Writer`-oriented methods
 - Support for arbitrary type system extensions
 - [Preprocessor directives](http://github.com/tinylib/msgp/wiki/Preprocessor-Directives)
 - File-based dependency model means fast codegen regardless of source tree size.

Consider the following:
```go
const Eight = 8
type MyInt int
type Data []byte

type Struct struct {
	Which  map[string]*MyInt `msg:"which"`
	Other  Data              `msg:"other"`
	Nums   [Eight]float64    `msg:"nums"`
}
```
As long as the declarations of `MyInt` and `Data` are in the same file as `Struct`, the parser will determine that the type information for `MyInt` and `Data` can be passed into the definition of `Struct` before its methods are generated.

#### Extensions

MessagePack supports defining your own types through "extensions," which are just a tuple of
the data "type" (`int8`) and the raw binary. You [can see a worked example in the wiki.](http://github.com/tinylib/msgp/wiki/Using-Extensions)

### Status

Mostly stable, in that no breaking changes have been made to the `/msgp` library in more than a year. Newer versions
of the code may generate different code than older versions for performance reasons. I (@philhofer) am aware of a
number of stability-critical commercial applications that use this code with good results. But, caveat emptor.

You can read more about how `msgp` maps MessagePack types onto Go types [in the wiki](http://github.com/tinylib/msgp/wiki).

Here some of the known limitations/restrictions:

- Identifiers from outside the processed source file are assumed (optimistically) to satisfy the generator's interfaces. If this isn't the case, your code will fail to compile.
- Like most serializers, `chan` and `func` fields are ignored, as well as non-exported fields.
- Encoding of `interface{}` is limited to built-ins or types that have explicit encoding methods.
- _Maps must have `string` keys._ This is intentional (as it preserves JSON interop.) Although non-string map keys are not forbidden by the MessagePack standard, many serializers impose this restriction. (It also means *any* well-formed `struct` can be de-serialized into a `map[string]interface{}`.) The only exception to this rule is that the deserializers will allow you to read map keys encoded as `bin` types, due to the fact that some legacy encodings permitted this. (However, those values will still be cast to Go `string`s, and they will be converted to `str` types when re-encoded. It is the responsibility of the user to ensure that map keys are UTF-8 safe in this case.) The same rules hold true for JSON translation.

If the output compiles, then there's a pretty good chance things are fine. (Plus, we generate tests for you.) *Please, please, please* file an issue if you think the generator is writing broken code.

### Performance

If you like benchmarks, see [here](http://bravenewgeek.com/so-you-wanna-go-fast/) and [here](https://github.com/alecthomas/go_serialization_benchmarks).

As one might expect, the generated methods that deal with `[]byte` are faster for small objects, but the `io.Reader/Writer` methods are generally more memory-efficient (and, at some point, faster) for large (> 2KB) objects.
