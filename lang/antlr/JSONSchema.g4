
grammar JSONSchema;

// Parser rules
schema
    : object EOF
    ;

object
    : '{' (pair (',' pair)*)? '}'
    ;

array
    : '[' (value (',' value)*)? ']'
    ;

pair
    : STRING ':' value
    ;

value
    : STRING
    | NUMBER
    | object
    | array
    | 'true'
    | 'false'
    | 'null'
    ;

// Lexer rules
STRING
    : '"' (~["\\\r\n] | escape_sequence)* '"'
    ;

NUMBER
    : '-'? ('0' | [1-9] [0-9]*) ('.' [0-9]+)? ([eE] [+-]? [0-9]+)?
    ;

fragment escape_sequence
    : '\\' (["\\/bfnrt] | 'u' [0-9a-fA-F] [0-9a-fA-F] [0-9a-fA-F] [0-9a-fA-F])
    ;

WHITESPACE
    : [ \t\r\n]+ -> skip
    ;
