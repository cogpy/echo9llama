
grammar Modelfile;

// Parser rules
modelfile
    : command* EOF
    ;

command
    : fromCommand
    | licenseCommand
    | templateCommand
    | systemCommand
    | adapterCommand
    | parameterCommand
    | messageCommand
    | comment
    | NEWLINE
    ;

fromCommand
    : FROM path
    ;

licenseCommand
    : LICENSE quotedString
    ;

templateCommand
    : TEMPLATE quotedString
    ;

systemCommand
    : SYSTEM quotedString
    ;

adapterCommand
    : ADAPTER path
    ;

parameterCommand
    : PARAMETER IDENTIFIER value
    ;

messageCommand
    : MESSAGE messageRole quotedString
    ;

messageRole
    : SYSTEM_ROLE
    | USER_ROLE
    | ASSISTANT_ROLE
    ;

quotedString
    : QUOTED_STRING
    | TRIPLE_QUOTED_STRING
    | UNQUOTED_STRING
    ;

path
    : UNQUOTED_STRING
    | QUOTED_STRING
    ;

value
    : UNQUOTED_STRING
    | QUOTED_STRING
    | NUMBER
    ;

comment
    : COMMENT
    ;

// Lexer rules
FROM : 'FROM' ;
LICENSE : 'LICENSE' ;
TEMPLATE : 'TEMPLATE' ;
SYSTEM : 'SYSTEM' ;
ADAPTER : 'ADAPTER' ;
PARAMETER : 'PARAMETER' ;
MESSAGE : 'MESSAGE' ;

SYSTEM_ROLE : 'system' ;
USER_ROLE : 'user' ;
ASSISTANT_ROLE : 'assistant' ;

TRIPLE_QUOTED_STRING
    : '"""' .*? '"""'
    ;

QUOTED_STRING
    : '"' (~["\r\n] | '\\"')* '"'
    ;

IDENTIFIER
    : [a-zA-Z_][a-zA-Z0-9_]*
    ;

NUMBER
    : [0-9]+ ('.' [0-9]+)?
    ;

UNQUOTED_STRING
    : ~[ \t\r\n"#]+
    ;

COMMENT
    : '#' ~[\r\n]* -> channel(HIDDEN)
    ;

NEWLINE
    : [\r\n]+ -> skip
    ;

WHITESPACE
    : [ \t]+ -> skip
    ;
