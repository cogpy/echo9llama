package scheme

import (
	"fmt"
	"strings"
	"sync"
)

// SchemeMetamodel provides symbolic reasoning capabilities through Scheme
// This is the Cognitive Grammar Kernel that enables meta-cognitive reflection
type SchemeMetamodel struct {
	mu          sync.RWMutex
	environment *Environment
	evaluator   *Evaluator
	parser      *Parser
	running     bool
}

// Environment represents the Scheme evaluation environment
type Environment struct {
	mu       sync.RWMutex
	bindings map[string]Value
	parent   *Environment
}

// Value represents a Scheme value
type Value interface {
	String() string
	Type() string
}

// Symbol represents a Scheme symbol
type Symbol struct {
	Name string
}

func (s *Symbol) String() string { return s.Name }
func (s *Symbol) Type() string   { return "symbol" }

// Number represents a Scheme number
type Number struct {
	Value float64
}

func (n *Number) String() string { return fmt.Sprintf("%v", n.Value) }
func (n *Number) Type() string   { return "number" }

// String represents a Scheme string
type String struct {
	Value string
}

func (s *String) String() string { return fmt.Sprintf("\"%s\"", s.Value) }
func (s *String) Type() string   { return "string" }

// Boolean represents a Scheme boolean
type Boolean struct {
	Value bool
}

func (b *Boolean) String() string {
	if b.Value {
		return "#t"
	}
	return "#f"
}
func (b *Boolean) Type() string { return "boolean" }

// Nil represents Scheme nil/empty list
type Nil struct{}

func (n *Nil) String() string { return "()" }
func (n *Nil) Type() string   { return "nil" }

// Cons represents a Scheme cons cell (pair)
type Cons struct {
	Car Value
	Cdr Value
}

func (c *Cons) String() string {
	var parts []string
	current := Value(c)

	for {
		if cons, ok := current.(*Cons); ok {
			parts = append(parts, cons.Car.String())
			current = cons.Cdr
		} else if _, ok := current.(*Nil); ok {
			break
		} else {
			// Improper list
			parts = append(parts, ".", current.String())
			break
		}
	}

	return "(" + strings.Join(parts, " ") + ")"
}
func (c *Cons) Type() string { return "cons" }

// Lambda represents a Scheme lambda function
type Lambda struct {
	Params []string
	Body   Value
	Env    *Environment
}

func (l *Lambda) String() string {
	return fmt.Sprintf("(lambda %v %v)", l.Params, l.Body)
}
func (l *Lambda) Type() string { return "lambda" }

// Primitive represents a built-in primitive function
type Primitive struct {
	Name string
	Fn   func(args []Value) (Value, error)
}

func (p *Primitive) String() string { return fmt.Sprintf("<primitive:%s>", p.Name) }
func (p *Primitive) Type() string   { return "primitive" }

// Parser parses Scheme expressions
type Parser struct {
	tokens []string
	pos    int
}

// Evaluator evaluates Scheme expressions
type Evaluator struct {
	mu sync.RWMutex
}

// NewSchemeMetamodel creates a new Scheme metamodel
func NewSchemeMetamodel() *SchemeMetamodel {
	sm := &SchemeMetamodel{
		environment: NewEnvironment(nil),
		evaluator:   &Evaluator{},
		parser:      &Parser{},
	}

	// Initialize primitive functions
	sm.initializePrimitives()

	return sm
}

// NewEnvironment creates a new environment
func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		bindings: make(map[string]Value),
		parent:   parent,
	}
}

// Start starts the Scheme metamodel
func (sm *SchemeMetamodel) Start() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.running {
		return fmt.Errorf("Scheme metamodel already running")
	}

	sm.running = true
	fmt.Println("🎭 Scheme Metamodel: Starting cognitive grammar kernel...")

	return nil
}

// Stop stops the Scheme metamodel
func (sm *SchemeMetamodel) Stop() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.running {
		return fmt.Errorf("Scheme metamodel not running")
	}

	sm.running = false
	fmt.Println("🎭 Scheme Metamodel: Stopping cognitive grammar kernel...")

	return nil
}

// Eval evaluates a Scheme expression
func (sm *SchemeMetamodel) Eval(expr string) (Value, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Parse expression
	value, err := sm.parser.Parse(expr)
	if err != nil {
		return nil, err
	}

	// Evaluate
	return sm.evaluator.Eval(value, sm.environment)
}

// Parse parses a Scheme expression string
func (p *Parser) Parse(expr string) (Value, error) {
	// Tokenize
	tokens := tokenize(expr)
	if len(tokens) == 0 {
		return &Nil{}, nil
	}

	p.tokens = tokens
	p.pos = 0

	return p.parseExpr()
}

// parseExpr parses a single expression
func (p *Parser) parseExpr() (Value, error) {
	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of input")
	}

	token := p.tokens[p.pos]
	p.pos++

	switch token {
	case "(":
		return p.parseList()
	case ")":
		return nil, fmt.Errorf("unexpected )")
	case "#t":
		return &Boolean{Value: true}, nil
	case "#f":
		return &Boolean{Value: false}, nil
	default:
		// Try to parse as number
		var num float64
		if _, err := fmt.Sscanf(token, "%f", &num); err == nil {
			return &Number{Value: num}, nil
		}

		// Try to parse as string
		if strings.HasPrefix(token, "\"") && strings.HasSuffix(token, "\"") {
			return &String{Value: token[1 : len(token)-1]}, nil
		}

		// Otherwise it's a symbol
		return &Symbol{Name: token}, nil
	}
}

// parseList parses a list
func (p *Parser) parseList() (Value, error) {
	var elements []Value

	for p.pos < len(p.tokens) && p.tokens[p.pos] != ")" {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		elements = append(elements, expr)
	}

	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("unclosed list")
	}

	p.pos++ // consume )

	// Build cons list
	return listToCons(elements), nil
}

// Eval evaluates a value in an environment
func (e *Evaluator) Eval(value Value, env *Environment) (Value, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch v := value.(type) {
	case *Symbol:
		// Variable lookup
		return env.Lookup(v.Name)

	case *Number, *String, *Boolean, *Nil, *Lambda, *Primitive:
		// Self-evaluating
		return value, nil

	case *Cons:
		// Function application or special form
		return e.evalCons(v, env)

	default:
		return nil, fmt.Errorf("unknown value type: %T", value)
	}
}

// evalCons evaluates a cons cell (function application or special form)
func (e *Evaluator) evalCons(cons *Cons, env *Environment) (Value, error) {
	// Check for special forms
	if sym, ok := cons.Car.(*Symbol); ok {
		switch sym.Name {
		case "quote":
			return e.evalQuote(cons, env)
		case "define":
			return e.evalDefine(cons, env)
		case "lambda":
			return e.evalLambda(cons, env)
		case "if":
			return e.evalIf(cons, env)
		case "begin":
			return e.evalBegin(cons, env)
		}
	}

	// Function application
	fn, err := e.Eval(cons.Car, env)
	if err != nil {
		return nil, err
	}

	// Evaluate arguments
	args, err := e.evalList(cons.Cdr, env)
	if err != nil {
		return nil, err
	}

	// Apply function
	return e.apply(fn, args)
}

// evalQuote evaluates a quote special form
func (e *Evaluator) evalQuote(cons *Cons, env *Environment) (Value, error) {
	if cons.Cdr == nil {
		return nil, fmt.Errorf("quote requires an argument")
	}

	if cdr, ok := cons.Cdr.(*Cons); ok {
		return cdr.Car, nil
	}

	return nil, fmt.Errorf("invalid quote form")
}

// evalDefine evaluates a define special form
func (e *Evaluator) evalDefine(cons *Cons, env *Environment) (Value, error) {
	if cons.Cdr == nil {
		return nil, fmt.Errorf("define requires arguments")
	}

	cdr, ok := cons.Cdr.(*Cons)
	if !ok {
		return nil, fmt.Errorf("invalid define form")
	}

	// Get variable name
	sym, ok := cdr.Car.(*Symbol)
	if !ok {
		return nil, fmt.Errorf("define requires a symbol")
	}

	// Get value
	if cdr.Cdr == nil {
		return nil, fmt.Errorf("define requires a value")
	}

	valueCons, ok := cdr.Cdr.(*Cons)
	if !ok {
		return nil, fmt.Errorf("invalid define form")
	}

	value, err := e.Eval(valueCons.Car, env)
	if err != nil {
		return nil, err
	}

	env.Define(sym.Name, value)
	return value, nil
}

// evalLambda evaluates a lambda special form
func (e *Evaluator) evalLambda(cons *Cons, env *Environment) (Value, error) {
	if cons.Cdr == nil {
		return nil, fmt.Errorf("lambda requires arguments")
	}

	cdr, ok := cons.Cdr.(*Cons)
	if !ok {
		return nil, fmt.Errorf("invalid lambda form")
	}

	// Get parameters
	params, err := extractParams(cdr.Car)
	if err != nil {
		return nil, err
	}

	// Get body
	if cdr.Cdr == nil {
		return nil, fmt.Errorf("lambda requires a body")
	}

	bodyCons, ok := cdr.Cdr.(*Cons)
	if !ok {
		return nil, fmt.Errorf("invalid lambda form")
	}

	return &Lambda{
		Params: params,
		Body:   bodyCons.Car,
		Env:    env,
	}, nil
}

// evalIf evaluates an if special form
func (e *Evaluator) evalIf(cons *Cons, env *Environment) (Value, error) {
	if cons.Cdr == nil {
		return nil, fmt.Errorf("if requires arguments")
	}

	cdr, ok := cons.Cdr.(*Cons)
	if !ok {
		return nil, fmt.Errorf("invalid if form")
	}

	// Evaluate condition
	cond, err := e.Eval(cdr.Car, env)
	if err != nil {
		return nil, err
	}

	// Check if true
	isTrue := true
	if b, ok := cond.(*Boolean); ok {
		isTrue = b.Value
	} else if _, ok := cond.(*Nil); ok {
		isTrue = false
	}

	if cdr.Cdr == nil {
		return nil, fmt.Errorf("if requires then branch")
	}

	thenCons, ok := cdr.Cdr.(*Cons)
	if !ok {
		return nil, fmt.Errorf("invalid if form")
	}

	if isTrue {
		return e.Eval(thenCons.Car, env)
	}

	// Else branch
	if thenCons.Cdr != nil {
		if elseCons, ok := thenCons.Cdr.(*Cons); ok {
			return e.Eval(elseCons.Car, env)
		}
	}

	return &Nil{}, nil
}

// evalBegin evaluates a begin special form
func (e *Evaluator) evalBegin(cons *Cons, env *Environment) (Value, error) {
	var result Value = &Nil{}
	var err error

	current := cons.Cdr
	for current != nil {
		if c, ok := current.(*Cons); ok {
			result, err = e.Eval(c.Car, env)
			if err != nil {
				return nil, err
			}
			current = c.Cdr
		} else {
			break
		}
	}

	return result, nil
}

// evalList evaluates a list of expressions
func (e *Evaluator) evalList(value Value, env *Environment) ([]Value, error) {
	var results []Value

	current := value
	for current != nil {
		if c, ok := current.(*Cons); ok {
			result, err := e.Eval(c.Car, env)
			if err != nil {
				return nil, err
			}
			results = append(results, result)
			current = c.Cdr
		} else if _, ok := current.(*Nil); ok {
			break
		} else {
			return nil, fmt.Errorf("invalid list")
		}
	}

	return results, nil
}

// apply applies a function to arguments
func (e *Evaluator) apply(fn Value, args []Value) (Value, error) {
	switch f := fn.(type) {
	case *Primitive:
		return f.Fn(args)

	case *Lambda:
		// Create new environment
		newEnv := NewEnvironment(f.Env)

		// Bind parameters
		if len(args) != len(f.Params) {
			return nil, fmt.Errorf("wrong number of arguments: expected %d, got %d",
				len(f.Params), len(args))
		}

		for i, param := range f.Params {
			newEnv.Define(param, args[i])
		}

		// Evaluate body
		return e.Eval(f.Body, newEnv)

	default:
		return nil, fmt.Errorf("not a function: %v", fn)
	}
}

// Define defines a variable in the environment
func (env *Environment) Define(name string, value Value) {
	env.mu.Lock()
	defer env.mu.Unlock()

	env.bindings[name] = value
}

// Lookup looks up a variable in the environment
func (env *Environment) Lookup(name string) (Value, error) {
	env.mu.RLock()
	defer env.mu.RUnlock()

	if value, ok := env.bindings[name]; ok {
		return value, nil
	}

	if env.parent != nil {
		return env.parent.Lookup(name)
	}

	return nil, fmt.Errorf("undefined variable: %s", name)
}

// initializePrimitives initializes built-in primitive functions
func (sm *SchemeMetamodel) initializePrimitives() {
	// Arithmetic
	sm.environment.Define("+", &Primitive{
		Name: "+",
		Fn: func(args []Value) (Value, error) {
			sum := 0.0
			for _, arg := range args {
				if num, ok := arg.(*Number); ok {
					sum += num.Value
				} else {
					return nil, fmt.Errorf("+ requires numbers")
				}
			}
			return &Number{Value: sum}, nil
		},
	})

	sm.environment.Define("-", &Primitive{
		Name: "-",
		Fn: func(args []Value) (Value, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("- requires at least one argument")
			}
			num, ok := args[0].(*Number)
			if !ok {
				return nil, fmt.Errorf("- requires numbers")
			}
			result := num.Value
			for i := 1; i < len(args); i++ {
				if num, ok := args[i].(*Number); ok {
					result -= num.Value
				} else {
					return nil, fmt.Errorf("- requires numbers")
				}
			}
			return &Number{Value: result}, nil
		},
	})

	// List operations
	sm.environment.Define("cons", &Primitive{
		Name: "cons",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("cons requires 2 arguments")
			}
			return &Cons{Car: args[0], Cdr: args[1]}, nil
		},
	})

	sm.environment.Define("car", &Primitive{
		Name: "car",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("car requires 1 argument")
			}
			if cons, ok := args[0].(*Cons); ok {
				return cons.Car, nil
			}
			return nil, fmt.Errorf("car requires a cons cell")
		},
	})

	sm.environment.Define("cdr", &Primitive{
		Name: "cdr",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("cdr requires 1 argument")
			}
			if cons, ok := args[0].(*Cons); ok {
				return cons.Cdr, nil
			}
			return nil, fmt.Errorf("cdr requires a cons cell")
		},
	})
}

// Helper functions

func tokenize(expr string) []string {
	// Simple tokenization
	expr = strings.ReplaceAll(expr, "(", " ( ")
	expr = strings.ReplaceAll(expr, ")", " ) ")
	return strings.Fields(expr)
}

func listToCons(elements []Value) Value {
	if len(elements) == 0 {
		return &Nil{}
	}

	var result Value = &Nil{}
	for i := len(elements) - 1; i >= 0; i-- {
		result = &Cons{Car: elements[i], Cdr: result}
	}

	return result
}

func extractParams(value Value) ([]string, error) {
	var params []string

	current := value
	for current != nil {
		if c, ok := current.(*Cons); ok {
			if sym, ok := c.Car.(*Symbol); ok {
				params = append(params, sym.Name)
				current = c.Cdr
			} else {
				return nil, fmt.Errorf("parameter must be a symbol")
			}
		} else if _, ok := current.(*Nil); ok {
			break
		} else {
			return nil, fmt.Errorf("invalid parameter list")
		}
	}

	return params, nil
}
