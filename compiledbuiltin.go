package mml

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

type Function struct {
	F         func([]interface{}) interface{}
	FixedArgs int
	args      []interface{}
}

func (f *Function) Bind(a []interface{}) *Function {
	b := *f
	b.args = a
	return &b
}

func (f *Function) Call(a []interface{}) interface{} {
	a = append(f.args, a...)
	if len(a) < f.FixedArgs {
		return f.Bind(a)
	}

	return f.F(a)
}

func Ref(v, k interface{}) interface{} {
	switch vt := v.(type) {
	case string:
		return string(vt[k.(int)])
	case []interface{}:
		return vt[k.(int)]
	case map[string]interface{}:
		return vt[k.(string)]
	default:
		panic("ref: unsupported code")
	}
}

func RefRange(v, from, to interface{}) interface{} {
	switch vt := v.(type) {
	case string:
		switch {
		case from == nil && to == nil:
			return vt[:]
		case from == nil:
			return vt[:to.(int)]
		case to == nil:
			return vt[from.(int):]
		default:
			return vt[from.(int):to.(int)]
		}
	case []interface{}:
		switch {
		case from == nil && to == nil:
			return vt[:]
		case from == nil:
			return vt[:to.(int)]
		case to == nil:
			return vt[from.(int):]
		default:
			return vt[from.(int):to.(int)]
		}
	default:
		panic("ref range: unsupported code")
	}
}

func UnaryOp(op int, arg interface{}) interface{} {
	switch unaryOperator(op) {
	case binaryNot:
		switch at := arg.(type) {
		case int:
			return +at
		default:
			panic("unary: unsupported code")
		}
	case plus:
		switch at := arg.(type) {
		case int:
			return +at
		case float64:
			return +at
		default:
			panic("unary: unsupported code")
		}
	case minus:
		switch at := arg.(type) {
		case int:
			return -at
		case float64:
			return -at
		default:
			panic("unary: unsupported code")
		}
	default:
		panic("unary: unsupported code")
	}
}

func BinaryOp(op int, left, right interface{}) interface{} {
	switch binaryOperator(op) {
	case binaryAnd:
		switch lt := left.(type) {
		case int:
			return lt & right.(int)
		default:
			panic("binary: unsupported code")
		}
	case binaryOr:
		switch lt := left.(type) {
		case int:
			return lt | right.(int)
		default:
			panic("binary: unsupported code")
		}
	case xor:
		switch lt := left.(type) {
		case int:
			return lt ^ right.(int)
		default:
			panic("binary: unsupported code")
		}
	case andNot:
		switch lt := left.(type) {
		case int:
			return lt &^ right.(int)
		default:
			panic("binary: unsupported code")
		}
	case lshift:
		switch lt := left.(type) {
		case int:
			return lt << right.(uint)
		default:
			panic("binary: unsupported code")
		}
	case rshift:
		switch lt := left.(type) {
		case int:
			return lt >> right.(uint)
		default:
			panic("binary: unsupported code")
		}
	case mul:
		switch lt := left.(type) {
		case int:
			return lt * right.(int)
		case float64:
			return lt * right.(float64)
		default:
			panic("binary: unsupported code")
		}
	case div:
		switch lt := left.(type) {
		case int:
			return lt / right.(int)
		case float64:
			return lt / right.(float64)
		default:
			panic("binary: unsupported code")
		}
	case mod:
		switch lt := left.(type) {
		case int:
			return lt % right.(int)
		default:
			panic("binary: unsupported code")
		}
	case add:
		switch lt := left.(type) {
		case int:
			return lt + right.(int)
		case float64:
			return lt + right.(float64)
		case string:
			return lt + right.(string)
		default:
			panic("binary: add: unsupported code")
		}
	case sub:
		switch lt := left.(type) {
		case int:
			return lt - right.(int)
		case float64:
			return lt - right.(float64)
		default:
			panic("binary: sub: unsupported code")
		}
	case eq:
		return left == right
	case notEq:
		return left != right
	case less:
		switch lt := left.(type) {
		case int:
			return lt < right.(int)
		case float64:
			return lt < right.(float64)
		case string:
			return lt < right.(string)
		default:
			panic("binary: less: unsupported code")
		}
	case lessOrEq:
		switch lt := left.(type) {
		case int:
			return lt <= right.(int)
		case float64:
			return lt <= right.(float64)
		case string:
			return lt <= right.(string)
		default:
			panic("binary: less-or-eq: unsupported code")
		}
	case greater:
		switch lt := left.(type) {
		case int:
			return lt > right.(int)
		case float64:
			return lt > right.(float64)
		case string:
			return lt > right.(string)
		default:
			panic("binary: greater: unsupported code")
		}
	case greaterOrEq:
		switch lt := left.(type) {
		case int:
			return lt >= right.(int)
		case float64:
			return lt >= right.(float64)
		case string:
			return lt >= right.(string)
		default:
			panic("binary: greater-or-eq: unsupported code")
		}
	default:
		panic("binary: unsupported code")
	}
}

func Nop(...interface{}) {}

var Len = &Function{
	F: func(a []interface{}) interface{} {
		switch at := a[0].(type) {
		case []interface{}:
			return len(at)
		case map[string]interface{}:
			return len(at)
		case string:
			return len(at)
		default:
			panic("len: unsupported code")
		}
	},
	FixedArgs: 1,
}

var IsError = &Function{
	F: func(a []interface{}) interface{} {
		_, ok := a[0].(error)
		return ok
	},
	FixedArgs: 1,
}

var Keys = &Function{
	F: func(a []interface{}) interface{} {
		s, ok := a[0].(map[string]interface{})
		if !ok {
			panic("keys: unsupported code")
		}

		var keys []interface{}
		for k := range s {
			keys = append(keys, k)
		}

		return keys
	},
	FixedArgs: 1,
}

var Format = &Function{
	F: func(a []interface{}) interface{} {
		f, ok := a[0].(string)
		if !ok {
			panic("format: unsupported code")
		}

		args, ok := a[1].([]interface{})
		if !ok {
			panic("format: unsupported code")
		}

		return fmt.Sprintf(f, args...)
	},
	FixedArgs: 2,
}

var Stderr = &Function{
	F: func(a []interface{}) interface{} {
		s, ok := a[0].(string)
		if !ok {
			panic("stderr: unsupported code")
		}

		_, err := os.Stderr.Write([]byte(s))
		return err
	},
	FixedArgs: 1,
}

var Stdout = &Function{
	F: func(a []interface{}) interface{} {
		s, ok := a[0].(string)
		if !ok {
			panic("stderr: unsupported code")
		}

		_, err := os.Stdout.Write([]byte(s))
		return err
	},
	FixedArgs: 1,
}

var Stdin = &Function{
	F: func(a []interface{}) interface{} {
		if a[0].(int) < 0 {
			b, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			return string(b)
		}

		b := make([]byte, a[0].(int))
		n, err := os.Stdin.Read(b)
		if err != nil {
			return err
		}

		return string(b[:n])
	},
	FixedArgs: 1,
}

var String = &Function{
	F: func(a []interface{}) interface{} {
		return fmt.Sprint(a[0])
	},
	FixedArgs: 1,
}

var Parse = &Function{
	F: func(a []interface{}) interface{} {
		code, err := parseModule(a[0].(string))
		if err != nil {
			return err
		}

		return codeCompiled(code)
	},
	FixedArgs: 1,
}

var Has = &Function{
	F: func(a []interface{}) interface{} {
		s, ok := a[1].(map[string]interface{})
		if !ok {
			return false
		}

		_, ok = s[a[0].(string)]
		return ok
	},
	FixedArgs: 2,
}

var IsBool = &Function{
	F: func(a []interface{}) interface{} {
		_, ok := a[0].(bool)
		return ok
	},
	FixedArgs: 1,
}

var IsInt = &Function{
	F: func(a []interface{}) interface{} {
		_, ok := a[0].(int)
		return ok
	},
	FixedArgs: 1,
}

var IsFloat = &Function{
	F: func(a []interface{}) interface{} {
		_, ok := a[0].(float64)
		return ok
	},
	FixedArgs: 1,
}

var IsString = &Function{
	F: func(a []interface{}) interface{} {
		_, ok := a[0].(string)
		return ok
	},
	FixedArgs: 1,
}

var Error = &Function{
	F: func(a []interface{}) interface{} {
		return errors.New(a[0].(string))
	},
	FixedArgs: 1,
}