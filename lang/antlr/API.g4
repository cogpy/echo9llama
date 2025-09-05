
grammar API;

// Parser rules
api_request
    : '{' request_field (',' request_field)* '}'
    ;

request_field
    : model_field
    | prompt_field
    | messages_field
    | parameters_field
    | stream_field
    | template_field
    | system_field
    ;

model_field
    : '"model"' ':' STRING
    ;

prompt_field
    : '"prompt"' ':' STRING
    ;

messages_field
    : '"messages"' ':' '[' (message (',' message)*)? ']'
    ;

message
    : '{' message_field (',' message_field)* '}'
    ;

message_field
    : role_field
    | content_field
    ;

role_field
    : '"role"' ':' STRING
    ;

content_field
    : '"content"' ':' STRING
    ;

parameters_field
    : '"parameters"' ':' object
    ;

stream_field
    : '"stream"' ':' BOOLEAN
    ;

template_field
    : '"template"' ':' STRING
    ;

system_field
    : '"system"' ':' STRING
    ;

object
    : '{' (pair (',' pair)*)? '}'
    ;

pair
    : STRING ':' value
    ;

value
    : STRING
    | NUMBER
    | BOOLEAN
    | 'null'
    | object
    | array
    ;

array
    : '[' (value (',' value)*)? ']'
    ;

// Lexer rules
STRING
    : '"' (~["\\\r\n] | escape_sequence)* '"'
    ;

NUMBER
    : '-'? ('0' | [1-9] [0-9]*) ('.' [0-9]+)? ([eE] [+-]? [0-9]+)?
    ;

BOOLEAN
    : 'true'
    | 'false'
    ;

fragment escape_sequence
    : '\\' (["\\/bfnrt] | 'u' [0-9a-fA-F] [0-9a-fA-F] [0-9a-fA-F] [0-9a-fA-F])
    ;

WHITESPACE
    : [ \t\r\n]+ -> skip
    ;
