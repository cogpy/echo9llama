
grammar Grammar;

// Parser rules
grammar_file
    : rule_definition* EOF
    ;

rule_definition
    : IDENTIFIER '::=' rule_body
    ;

rule_body
    : alternation
    ;

alternation
    : concatenation ('|' concatenation)*
    ;

concatenation
    : element+
    ;

element
    : atom quantifier?
    ;

atom
    : IDENTIFIER                    // rule reference
    | STRING_LITERAL               // terminal string
    | character_class              // character class
    | '(' alternation ')'          // grouping
    | '.'                          // any character
    | '(?=' alternation ')'        // positive lookahead
    | '(?!' alternation ')'        // negative lookahead
    | '\\b'                        // word boundary
    | '\\B'                        // non-word boundary
    ;

character_class
    : '[' character_range+ ']'
    | '[' '^' character_range+ ']'  // negated character class
    ;

character_range
    : CHARACTER
    | CHARACTER '-' CHARACTER       // range
    | escape_sequence
    ;

escape_sequence
    : '\\' ["\\/bfnrt]
    | '\\' 'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT
    | '\\' .
    ;

quantifier
    : '?'                          // optional
    | '*'                          // zero or more
    | '+'                          // one or more
    | '{' NUMBER '}'               // exact count
    | '{' NUMBER ',' '}'           // minimum count
    | '{' NUMBER ',' NUMBER '}'    // range count
    ;

// Lexer rules
IDENTIFIER
    : [a-zA-Z_][a-zA-Z0-9_-]*
    ;

STRING_LITERAL
    : '"' (~["\\\r\n] | escape_sequence)* '"'
    ;

CHARACTER
    : ~[\\\]\-\r\n]
    | escape_sequence
    ;

NUMBER
    : [0-9]+
    ;

HEX_DIGIT
    : [0-9a-fA-F]
    ;

COMMENT
    : '#' ~[\r\n]* -> channel(HIDDEN)
    ;

WHITESPACE
    : [ \t\r\n]+ -> skip
    ;
