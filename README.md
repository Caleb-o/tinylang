# TinyLang
A small interpreted language to try and learn more about language design and development. TL is a tree-walk interpreter (for now) with a Rust/Pascal/Python-like feel.

***Note:** This language is very young and does not have everything a modern language may have. It is also not intended for use, other than for learning.*

## Contents
* [DataTypes](#data-types)
* [Examples](#examples)
* [Sample Scripts](#sample-scripts)
* [Builtin Functions](#builtin-functions)

## TODO

### Features
* Member access/assignment for lists and struct instances
	* list[0] / list[0] = 10
	* instance.field / instance.field = 10
* Imports - Imports a file in raw form (no wrapping in a namespace)
* Namespaces - Define your own namespaces to contain code
	* The main program may automatically use a Program namespace
* Inject functions, structs and namespaces from C# code (for native libraries)
* Builtin Data structures (with literals)
	* Dictionary - {"foo": 123, "bar": 456}

### Fixes / Modifications
* Nothing of interest right now :^)

## Desirables
* Bytecode interpreter
* Transpilation - Python and/or C++

## Inspirations
* Rust
* Zig
* Pascal
* Python

## Data Types
* int
* float
* bool
* string
* unit (to signify no return)
* list
* struct

## Examples

### Hello, World!
```julia
# This is a keyword masked as a function, so it looks more natural. It accepts any amount of arguments.
print("Hello, World!");
```

### Variables and Mutability
```julia
# A mutable variable (type is inferred)
var foo = 10;
foo = foo - 20;

# An immutable variable
let bar = foo * foo;

# Error: Cannot mutate an immutable variable
# bar = 20;

# Using a type annotation
var baz: int = 22;
```

### Functions
```julia
let simple = function() {
	# unit is the assumed return type when the return type is omitted
	print("Hello!");
};

let return_integer = function(): int {
	# Notice there is no return statement here, as an implicit
	# 'result' variable is made with the type int
	# and will be returned on function exit
	# The default value will be the default of a primitive,
	# otherwise it's a unit
};

simple(); # Hello!
let foo = return_integer();

print(foo, " ", 1234); # 0 1234
```

### Nested Functions
```julia
# Functions can be defined and called within each other
# They are scope based, so they can only be called from
# its current scope
let foo = function() {
	let bar = function() {
		let baz = function() {

		};

		baz();
	};

	bar();
};

foo();
```

### Control Flow
```julia
let a = 10;

# If statements
if a > 20 {
	print("A > 20");
} else if (a <= 10) {
	print("a <= 10");
} else {
	print("other");
}

# Looping with while loop
var i = 0;

while i < 10 {
	i = i + 1;
	print(i);
}

# -- Declare a variable within the while statement
while var j = 0; j < 10 {
	j = j + 1;
	print(j);
}

# -- Do While loop
i = 0;

do {
	i = i + 1;
} while i < 2;
```

## Sample Scripts

### [Fibonacci](./examples/fibonacci.tiny) (Recursive)
```julia
# Recursive function to get the Nth value of the fibonacci sequence
let fib = function(n: int): int {
	if n > 1 {
		result = fib(n - 1) + fib(n - 2);
	} else {
		result = n;
	}
};

print(fib(24)); # 46368
```

### [Fibonacci](./examples/fibonacci2.tiny) (Variable Swaps)
```julia
# This approach improves on performance dramatically. We can get a higher
# Nth value of the sequence, in a fraction of the time.
# This has to do with the performance of recursion
let fib = function(nth: int): int {
	var a = 0, b = 1, c = 0;

	if (nth == 0) {
		result = nth;
		return;
	}

	while var i = 2; i <= nth {
		i = i + 1;

		c = a + b;
		a = b;
		b = c;
	}

	result = b;
};

print(fib(32)); # 2178309
```

# Builtin Functions

## IO

### file_exists
* Parameters: file_name (string)
* Returns: bool

```julia
if file_exists("sample.txt") {
	print("File Exists");
}
```

### file_read
* Parameters: file_name (string)
* Returns: string

```julia
let content = file_read("sample.txt");
```

### file_write
* Parameters: file_name (string), content (string)
* Returns: unit

```julia
file_write("sample.txt", "Hello, World!");
```

## Misc

### eval
* Parameters: source (string)
* Returns: any

```julia
eval("result = 1 + 2;");
```

### types_match
* Parameters: left (any), right (any)
* Returns: bool

```julia
types_match(1, "string");
```

### type_of
* Parameters: value (any)
* Returns: string

```julia
type_of("string");
```