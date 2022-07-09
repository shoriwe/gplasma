- Reformatted lexer package
- Reformatted parser package
- Adoption of Go error handling
- Parser now parses if and unless statement elif blocks
- Pass to validate return is only inside functions and generator definitions
- Pass to validate yield is only inside generator definitions
- Pass to validate break/continue/redo is only inside loops
- New lines handled reader side
- New `block` statement
- New `delete` statement
- New `require` statement
- New `super` expression
- New `is` and `implements` binary operators
- Directory for test samples
- Simplify AST pass