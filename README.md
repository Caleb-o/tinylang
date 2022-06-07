# TinyLang
A small interpreted language to try and learn more about language design and development. TL is a tree-walk interpreter (for now) with a Rust/Pascal/Python-like feel.

***Note:** This language is very young and does not have everything a modern language may have. It is also not intended for use, other than for learning.*

## Contents
* [DataTypes](#data-types)
* [Examples](#examples)
* [Sample Scripts](#sample-scripts)
* [Builtins](#builtin-functions)

## Other
* [Caveats](./CAVEATS.md)

## TODO

### Features
* Structures - User types
* Enums - Either C-like or Rust-like
* Traits/Interfaces - Add "generic" functionality to structs eg. Hashable requiring a hash function
	* Pass-by-trait values
* Builtin Data structures (with literals)
	* List - [1, 2, 3]
	* Dictionary - {"foo": 123, "bar": 456}
* Ranges - 0..100, 0..=100
* Imports (Lazy?) - Import other scripts lazily, so they only get included once a call to the import is made

### Fixes / Modifications
* Nothing of interest right now :^)

## Desirables
* Bytecode interpreter
* Transpilation - Python and/or C++

## Inspirations
* Rust
* Pascal
* Python

## Data Types
* int
* float
* Bool
* string
* void (to signify no return)

## Examples

### Hello, World!
```julia
# This is a builtin function, which are called by prefixing 
# a function call with an @. They are defined in the VM.
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
	# void is the assumed return type when the return type is omitted
	@println("Hello!");
}

fn return_integer(): int {
	# Notice there is no return statement here, as an implicit
	# 'result' variable is made with the type int
	# and will be returned on function exit
}

fn return_default_integer(): int(1234) {
	# Functions can define a default return value
	# It is a niche feature, but feels right to have
}

simple(); # Hello!
let int foo = return_integer();
let int bar = return_default_integer();

@println(foo, " ", bar); # 0 1234
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

### Control Flow
```julia
let int a = 10;

# If statements
if a > 20 {
	@println("A > 20");
} else if (a <= 10) {
	@println("a <= 10);
} else {
	@println("other");
}

# Looping with while loop
var int i = 0;

while i < 10 {
	i = i + 1;
	@println(i);
}

# -- Declare a variable within the while statement
while var int j = 0; j < 10 {
	j = j + 1;
	@println(j);
}

# -- Do While loop
i = 0;

do {
	i = i + 1;
} while i < 2;
```

### Referencing Variables
```julia
# By default, function parameters are immutable, so they cannot
# be mutated. If they are marked with var, they then become mutable
# Only mutable variables can be passed in, as it wouldn't make sense
# to mutate a literal

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

## Sample Scripts

### [Fibonacci](./examples/fibonacci.tiny) (Recursive)
```julia
# Recursive function to get the Nth value of the fibonacci sequence
fn fibonacci(n: int): int {
	if n > 1 {
		result = fibonacci(n - 1) + fibonacci(n - 2);
	} else {
		result = n;
	}
}

@println(fibonacci(24)); # 46368
```

### [Fibonacci](./examples/fibonacci2.tiny) (Variable Swaps)
```julia
fn fibonacci(n: int): int {
	var int a = 0, b = 1, c = 0;

	if (n == 0) {
		result = n;
		return;
	}

	while var int i = 2; i <= n {
		i = i + 1;

		c = a + b;
		a = b;
		b = c;
	}

	result = b;
}

@println(fibonacci(24)); # 46368
```

## Builtin Functions
* println(...) : Variadic function to print values to the console
* printobj(...) : Variadic function to print more information about values