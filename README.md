# TinyLang
A small interpreted language to try and learn more about language design and development. TL is a tree-walk interpreter (for now) with a Rust/Pascal/Python-like feel.

***Note:** This language is very young and does not have everything a modern language may have. It is also not intended for use, other than for learning.*

## TODO
* Conditional - if/else/elif
* Loops - while/for
* Structures - User types
* Enums - Either C-like or Rust-like
* Traits/Interfaces - Add "generic" functionality to structs eg. Hashable requiring a hash function
	* Pass-by-trait values
* Builtin Data structures (with literals)
	* List - [1, 2, 3]
	* Dictionary - {"foo": 123, "bar": 456}
* Ranges - 0..100, 0..=100

## Desirables
* Bytecode interpreter
* Transpilation - Python and/or C++

## Inspirations
* Rust
* Pascal
* Python

## Examples

### Hello, World!
```julia
# This is a builtin function, which are called by prefixing 
# an function call with an @. They are defined in the VM.
@println("Hello, World!");
```

### Variables and Mutability
```julia
# A mutable variable
var int foo = 10;
foo = foo - 20;

# An immutable variable
let int bar = foo * foo;

# Error: Cannot mutate an immutable variable
# bar = 20;
```

### Functions
```julia
fn simple() {
	@println("Hello!");
}

fn return_integer(): int {
	# Notice there is no return as an implicit
	# 'result' variable is made with the type int
	# and will be returned on function exit
}

fn return_explicit_integer(): int(1234) {
	# Functions can define a default return value
}

simple();
let int foo = return_integer();
let int bar = return_explicit_integer();

@println(foo, " ", bar);
```

### Nested Functions
```julia
# Functions can be defined and called within each other
# They are scope based, so they can only be called from
# its current scope
fn foo() {
	fn bar() {
		fn baz() {

		};

		baz();
	};

	bar();
}

foo();
```

### Referencing Variables
```julia
# By default, function parameters are immutable, so they cannot
# be mutated. If they are marked with var, they then become mutable
# Only mutable variables can be passed in
# Note: This also works with nested functions
fn increment_reference(my_ref: var int) {
	# my_ref = my_var
	my_ref = my_ref + 1;
}

var int my_var = 0;
increment_reference(my_var); # 1
increment_reference(my_var); # 2
increment_reference(my_var); # 3

@println(my_var); # 3
```

### Value Capturing
```julia
# Since nested functions exists, it would be easier if we could
# modify values in outer scopes. If a variable being modified 
# doesn't exist in its own scope, it will climb up to find a variable with that name
var int a = 10;

fn foo() {
	fn bar() {
		fn baz() {
			a = 10;
		};

		baz();
	}

	bar();
}

foo();
@println(a); # 10
```

## Builtin Functions
* println(...) : Variadic function to print values to the console