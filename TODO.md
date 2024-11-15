# TODO

## journal file parser
- directives
    - [x] support account directive (`account`)
    - [x] support payee directive (`payee`)
    - [ ] support tag directive (`tag`)
    - [ ] support commodity directive (`commodity`)
    - [ ] support alias directive (`alias`)
    - [ ] support decimal-mark directive (`decimal-mark`)
    - [ ] support include directive (`include`)
    - [ ] support market price directive (`P`)
    - [ ] support default commodity directive (`D`)
    - [ ] support default year directive (`Y`)
    - [ ] support prepend account directive (`apply account` and `end apply account`)
- transactions
    - [x] support basic transaction lines
    - [ ] support recurring transactions (`~`)
    - [ ] support auto-posted transactions (`=`)
    - [ ] support inline comment tagged transactions
    - postings
        - [ ] support basic postings
        - [ ] support virtual postings (`[...]`)
        - [ ] support unbalanced postings (`(...)`)
        - [ ] support assertions (`= ...`)
        - [ ] support inline comment tagged postings
- comments
    - [x] support comments starting with `;` or `#`
    - [ ] support comments starting with `*`
    - [x] support inline comments starting with `  ;` or `  #`
    - [ ] support block comments (`comment` and `end comment`)
    - [ ] support tags in inline comments
    - [ ] support indented additional comments
- integration
    - [ ] support finding an AST token by position in a file

## language server
- [ ] code completion for account names
    - [ ] based on prefix-syntax, e.g. `exp:Ca:Che` should suggest `expenses:Cash:Checking`, if it exists
