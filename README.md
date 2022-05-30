# TinyLang
A small interpreted language because I can

## Example
```
record Person {
	var name;
}

fn say_hello(person) {
	print("Hello, ", person.name);
}

# Top-level code is fine because it's interpreted
var p = Person.new({ "Caleb" });
say_hello(p);
```

## Builtin Functions

### Type Checking
* is_num(n): bool - Returns if a value/variable is a number
* is_int(n): bool - Returns if a value/variable is a integer
* is_float(n): bool - Returns if a value/variable is a float
* is_string(n): bool - Returns if a value/variable is a string
* is_bool(n): bool - Returns if a value/variable is a bool
* is_record(n): bool - Returns if a value/variable is a record