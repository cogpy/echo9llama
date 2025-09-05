
grammar ThinkingParser;

// Parser rules
content
    : (thinkingBlock | regularContent)* EOF
    ;

thinkingBlock
    : openingTag thinkingContent closingTag
    ;

thinkingContent
    : (TEXT | WHITESPACE)*
    ;

regularContent
    : (TEXT | WHITESPACE)+
    ;

openingTag
    : OPENING_TAG
    ;

closingTag
    : CLOSING_TAG
    ;

// Lexer rules
OPENING_TAG
    : '<' 'thinking' '>'
    ;

CLOSING_TAG
    : '</' 'thinking' '>'
    ;

TEXT
    : ~[<\r\n\t ]+
    | '<' ~[/]
    | '</' ~[t]
    | '</t' ~[h]
    | '</th' ~[i]
    | '</thi' ~[n]
    | '</thin' ~[k]
    | '</think' ~[i]
    | '</thinki' ~[n]
    | '</thinkin' ~[g]
    | '</thinking' ~[>]
    ;

WHITESPACE
    : [ \t\r\n]+
    ;
