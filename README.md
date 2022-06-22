# TinyLang
A small interpreted language to try and learn more about language design and development. TL is a tree-walk interpreter (for now) with a Rust/Pascal/Python-like feel.

***Note:** This language is very young and does not have everything a modern language may have. It is also not intended for use, other than for learning.*

## Contents
* [DataTypes](#data-types)
* [Examples](#examples)
* [Sample Scripts](#sample-scripts)
* [Builtin Functions](#builtin-functions)

## TODO
* Imports
	* Import files into their own namespaces
	* Store result of imported file into a new representation
	* Analyse each new file seperately (index ASTs for analysis)
* Analysis
	* Resolve identifiers that aren't yet defined to allow calls before definitions
* Native functions, structs and namespaces from Go code (for native libraries)
* Builtin Data structures (with literals)
	* Lists - [1, 2, 3]
	* Dictionary - {"foo": 123, "bar": 456}
* "strict" modifier for variables
	* Cannot re-assign with a different type, acts like static type

### Features
* Dynamic typing
* Object Oriented systems
* First-class and higher-order functions
* Anonymous functions
* Exception-like throw/catch
	* Throw values and unwind until caught
* Loose immutability (disallow rebinding symbol, but not get/set of objects)
	* Note: Stricter mutability may come later

### Fixes / Modifications
* Nothing of interest right now :^)

## Desirables
* Bytecode interpreter
* Transpilation - Python, JavaScript and/or C++

## Inspirations
* Rust
* Zig
* Pascal
* Python
* Go

## Data Types
* int
* float
* bool
* string
* unit (to signify no return)
* class

## Examples

### Hello, World!
```coffee
# This is a keyword masked as a function, so it looks more natural. It accepts any amount of arguments.
print("Hello, World!");
```

### Variables
```coffee
# A variable
var foo = 10;
foo = foo - 20;
```

### Functions
```coffee
function simple() {
	# unit is the assumed return type when the return type is omitted
	print("Hello!");
}

simple(); # Hello!
```

### Nested Functions
```coffee
# Functions can be defined and called within each other
# They are scope based, so they can only be called from
# its current scope
function foo() {
	function bar() {
		function baz() {

		}

		baz();
	}

	bar();
}

foo();
```

### Control Flow
```coffee
var a = 10;

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
```

## Sample Scripts

### [Fibonacci](./examples/fibonacci.tiny): Recursive
```coffee
# Recursive function to get the Nth value of the fibonacci sequence
function fib(n) {
	if n > 1 {
		return fib(n - 1) + fib(n - 2);
	} else {
		return n;
	}
}

print(fib(24)); # 46368
```

### [Fibonacci](./examples/fibonacci2.tiny): Variable Swaps
```coffee
# This approach improves on performance dramatically. We can get a higher
# Nth value of the sequence, in a fraction of the time.
# This has to do with the performance of recursion
function fib(nth) {
	var a = 0;
	var b = 1;
	var c = 0;

	if (nth == 0) {
		return nth;
	}

	while var i = 2; i <= nth {
		i = i + 1;

		c = a + b;
		a = b;
		b = c;
	}

	return b;
}

print(fib(32)); # 2178309
```

### [Fibonacci](./examples/fibonacci3.tiny): Iterator-like with classes
```coffee
# Iterative fibonacci through the use of classes
class Fibonacci {
	var old;
	var value;
	
	function Fibonacci() {
		self.old = 0;
		self.value = 1;
	}

	function get() {
		return self.value;
	}

	function next() {
		var temp = self.old + self.value;
		self.old = self.value;
		self.value = temp;
	}
}

var fib = Fibonacci();

while var idx = 2; idx <= 32 {
	idx = idx + 1;
	fib.next();
}

print(fib.get()); # 2178309
```

# Builtin Functions
***N/A***