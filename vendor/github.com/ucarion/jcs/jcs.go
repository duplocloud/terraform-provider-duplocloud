// Package jcs implements JSON Canonicalization Scheme, formalized as RFC8785:
//
// https://www.rfc-editor.org/rfc/rfc8785.html
package jcs

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"unicode/utf16"
)

// ErrNaN indicates that the inputted value is or contains NaN.
var ErrNaN = errors.New("jcs: cannot c14n NaN")

// ErrInf indicates that the inputted value is or contains infinity or negative
// infinity.
var ErrInf = errors.New("jcs: cannot c14n Inf")

// ErrUnsupportedType indicates the inputted value is or contains a type outside
// the set of types described in Append.
var ErrUnsupportedType = errors.New("jcs: value has unsupported type")

// Format returns the JCS canonical representation of the given value.
//
// See Append for the types of data supported, and possible error conditions.
func Format(v interface{}) (string, error) {
	var b []byte
	b, err := Append(b, v)
	return string(b), err
}

type keyVal struct {
	key []uint16
	val interface{}
}

// Append appends the JCS canonical representation of the value v to b and
// returns the extended buffer.
//
// Append only supports the types used by encoding/json to unmarshal into an
// interface value, namely:
//
//  bool, for JSON booleans
//  float64, for JSON numbers
//  string, for JSON strings
//  []interface{}, for JSON arrays
//  map[string]interface{}, for JSON objects
//  nil for JSON null
//
// Any other types, including types embedded within maps or arrays in the value
// v, will result in Append returning ErrUnsupportedType.
//
// Furthermore, if v contains float64 values that are NaN, Infinity, or negative
// Infinity, then Append will return either ErrNaN or ErrInf.
//
// Append does not check if strings contain invalid surrogate pairs. Codepoints
// in strings will be appended to b as-is, with the appropriate escaping
// required by JCS.
func Append(b []byte, v interface{}) ([]byte, error) {
	// Handle the nil case right away here. encoding/json converts any type of nil
	// to JSON null.
	//
	// From the spec:
	//
	// In accordance with JSON [RFC8259], the literals "null", "true", and "false"
	// MUST be serialized as null, true, and false, respectively.
	if v == nil {
		b = append(b, 'n', 'u', 'l', 'l')
		return b, nil
	}

	switch v := v.(type) {
	case bool:
		// From the spec:
		//
		// In accordance with JSON [RFC8259], the literals "null", "true", and
		// "false" MUST be serialized as null, true, and false, respectively.
		if v {
			b = append(b, 't', 'r', 'u', 'e')
		} else {
			b = append(b, 'f', 'a', 'l', 's', 'e')
		}

		return b, nil
	case float64:
		// We handle the NaN and Inf cases here, and return errors in that case.
		//
		// From the spec:
		//
		// Note: Since Not a Number (NaN) and Infinity are not permitted in JSON,
		// occurrences of NaN or Infinity MUST cause a compliant JCS implementation
		// to terminate with an appropriate error.
		if math.IsNaN(v) {
			return nil, ErrNaN
		}

		if math.IsInf(v, 0) {
			return nil, ErrInf
		}

		// Because of negative zero case (which we must encode as "0", not "-0", but
		// strconv will want to encode "-0"), it's easiest to handle the zero cases
		// here.
		if v == 0 {
			b = append(b, '0')
			return b, nil
		}

		// The spec requires that we encode numbers like JavaScript does:
		//
		// ECMAScript builds on the IEEE 754 [IEEE754] double-precision standard for
		// representing JSON number data. Such data MUST be serialized according to
		// Section 7.1.12.1 of [ECMA-262], including the "Note 2" enhancement.
		//
		// This is pretty difficult to do completely correctly. Around the web, some
		// approaches are:
		//
		// https://github.com/cyberphone/json-canonicalization/blob/77e15c5d8e476c36081ffb9cfacab1c4336b6152/go/src/webpki.org/jsoncanonicalizer/es6numfmt.go
		//
		// Rundgren's Golang implementation of JCS, which is a slightly more
		// complicated implementation than this one, attempting to handle a few more
		// corner cases between strconv and the ECMAScript standard.
		//
		// At the time of writing, the standard test cases don't seem to elucidate
		// why all of the special cases in Rundgren's code are required.
		//
		// https://github.com/cespare/ryu
		//
		// The JCS RFC offers Ryu as a possible implementation, and Ryu is used in
		// the Java reference implementation of JCS. But the Golang implementation
		// of Ryu, at the time of writing, seeks compatibility with strconv, and
		// does not offer settings to force the e/f/g formats.
		//
		// https://github.com/dop251/goja/blob/52cab25ecbdd1c2cd42cd8bf904fa5b466b680f3/value.go#L501-L520
		//
		// Goja, a Golang implementation of JavaScript, takes an approach similar to
		// Rundgren's, but again simplified to just use strconv's 'g' or 'f' with a
		// regex to handle converting 'e+07' to 'e+7'.
		//
		// Ultimately, this code takes inspiration mostly from Rundgren and goja. We
		// look at the value of the number, and either use strconv's 'f' or 'g'
		// format, with the 'g' format trying to detect and strip away a leading
		// zero in the exponent part of the result.
		abs := math.Abs(v)
		if abs < 1e+21 && abs >= 1e-6 {
			b = strconv.AppendFloat(b, v, 'f', -1, 64)
		} else {
			// When we output with the exponential format, we're going to need to try
			// to find a substring like 'e+0' or 'e-0', and replace it with 'e+' or
			// 'e-'. This is to handle the fact that strconv may return:
			//
			// 9.999999999999997e-07
			//
			// But we require:
			//
			// 9.999999999999997e-7
			//
			// To do this, we'll work in a separate buffer, and then append to the
			// main buffer once we're doing doing our manipulations.

			var bb []byte
			bb = strconv.AppendFloat(bb, v, 'g', -1, 64)

			e := bytes.LastIndexByte(bb, 'e')
			if e >= 0 && bb[e+2] == '0' {
				// We need to splice out the '0' at index e+2 from bb. We can do this
				// without a further buffer, let's just add the sub-slices of bb into b.
				b = append(b, bb[:e+2]...)
				b = append(b, bb[e+3:]...)
			} else {
				// No splicing is required, we can just add bb to b
				b = append(b, bb...)
			}
		}

		return b, nil
	case string:
		b = appendString(b, []rune(v))
		return b, nil
	case []interface{}:
		// The spec does not explicitly call out how to handle arrays, beyond this
		// note:
		//
		// Whitespace between JSON tokens MUST NOT be emitted.
		//
		// There isn't much ambiguity as to how to handle arrays, so it's fine that
		// the spec doesn't elaborate here.
		b = append(b, '[')
		for i, vv := range v {
			var err error
			b, err = Append(b, vv)
			if err != nil {
				return nil, err
			}

			if i != len(v)-1 {
				b = append(b, ',')
			}
		}

		b = append(b, ']')
		return b, nil
	case map[string]interface{}:
		// The spec requires that object keys be sorted. In particular:
		//
		// When a JSON object is about to have its properties sorted, the following
		// measures MUST be adhered to:
		//
		// - The sorting process is applied to property name strings in their "raw"
		// (unescaped) form. That is, a newline character is treated as U+000A.
		//
		// - Property name strings to be sorted are formatted as arrays of UTF-16
		// [UNICODE] code units. The sorting is based on pure value comparisons,
		// where code units are treated as unsigned integers, independent of locale
		// settings.
		//
		// - Property name strings either have different values at some index that
		// is a valid index for both strings, or their lengths are different, or
		// both. If they have different values at one or more index positions, let k
		// be the smallest such index; then, the string whose value at position k
		// has the smaller value, as determined by using the "<" operator,
		// lexicographically precedes the other string. If there is no index
		// position at which they differ, then the shorter string lexicographically
		// precedes the longer string.
		//
		// In other words: sort the keys according to their UTF-16 []uint16
		// encodings. This corresponds to how JavaScript represents strings.
		pairs := make([]keyVal, 0, len(v))
		for k, v := range v {
			pairs = append(pairs, keyVal{key: utf16.Encode([]rune(k)), val: v})
		}

		// Sort the members of the object by their keys, breaking ties in favor of
		// shorter strings.
		sort.Slice(pairs, func(i, j int) bool {
			a := pairs[i].key
			b := pairs[j].key

			for i := 0; i < len(a) && i < len(b); i++ {
				if a[i] < b[i] {
					return true
				}

				if a[i] > b[i] {
					return false
				}
			}

			return len(a) < len(b)
		})

		b = append(b, '{')
		for i, pair := range pairs {
			b = appendString(b, utf16.Decode(pair.key))
			b = append(b, ':')

			var err error
			b, err = Append(b, pair.val)
			if err != nil {
				return nil, err
			}

			if i != len(pairs)-1 {
				b = append(b, ',')
			}
		}

		b = append(b, '}')
		return b, nil
	}

	// Any other type not handled already isn't supported. Ignoring this case (for
	// instance, by appending nothing) could lead to invalid JSON and might lead
	// people to mistakenly sign incomplete data. More prudent is to return an
	// error.
	return nil, ErrUnsupportedType
}

func appendString(b []byte, s []rune) []byte {
	// From the spec:
	//
	// For JSON string data (which includes JSON object property names as well),
	// each Unicode code point MUST be serialized as described below (see Section
	// 24.3.2.2 of [ECMA-262]):
	//
	// - If the Unicode value falls within the traditional ASCII control character
	// range (U+0000 through U+001F), it MUST be serialized using lowercase
	// hexadecimal Unicode notation (\uhhhh) unless it is in the set of predefined
	// JSON control characters U+0008, U+0009, U+000A, U+000C, or U+000D, which
	// MUST be serialized as \b, \t, \n, \f, and \r, respectively.
	//
	// - If the Unicode value is outside of the ASCII control character range, it
	// MUST be serialized "as is" unless it is equivalent to U+005C (\) or U+0022
	// ("), which MUST be serialized as \\ and \", respectively. Finally, the
	// resulting sequence of Unicode code points MUST be enclosed in double quotes
	// (").
	//
	// Note: Since invalid Unicode data like "lone surrogates" (e.g., U+DEAD) may
	// lead to interoperability issues including broken signatures, occurrences of
	// such data MUST cause a compliant JCS implementation to terminate with an
	// appropriate error.
	//
	// We here diverge from the spec, in that we do not attempt to detect and
	// error on lone surrogates, because doing so is rather difficult to do
	// correctly. Otherwise, we straightforwardly comply with the requirements.

	b = append(b, '"')
	for _, c := range s {
		if c == '\\' {
			b = append(b, '\\', '\\')
		} else if c == '"' {
			b = append(b, '\\', '"')
		} else if c == '\b' {
			b = append(b, '\\', 'b')
		} else if c == '\f' {
			b = append(b, '\\', 'f')
		} else if c == '\n' {
			b = append(b, '\\', 'n')
		} else if c == '\r' {
			b = append(b, '\\', 'r')
		} else if c == '\t' {
			b = append(b, '\\', 't')
		} else if c < 0x20 {
			b = append(b, fmt.Sprintf("\\u%04x", c)...)
		} else {
			b = append(b, string(c)...)
		}
	}

	b = append(b, '"')
	return b
}
