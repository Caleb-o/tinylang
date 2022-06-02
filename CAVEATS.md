# Caveats

## Uncached Literals
Currently literals are not cached and values are allocated every time they are evaluated. This is a *major* issue, as these values are very short-lived and a waste of an allocation. Example:

Simple binary expression
```
let int foo = 10 + 10;
```

This incredibly simple expression generates 3 values: 10, 10 and the result of expression. So 3 values are generated for only 1 to be used. The 10 should be cached in a constant table, so a value must be looked up instead of being generated on evaluation. If a constant table was used instead, there would only be 2 values: The 10 referenced twice and the result. This seems like a small optimisation, however, if you throw the exact expression into a loop, it will be *much* worse.

```
var int i = 0;

while i < 10 {
	i = i + 1;
	var int foo = 10 + 10;
}
```

Disregarding the index incrementing, the same result is here. We now have 3 values being generated, but now for every iteration. The total (without the index variable) is around 300 values generated. With the constants table, only 11 values would be generated. The initial value for the 10 literal and the result of the expression 10 times. This is a *HUGE* difference and reduction of allocations.