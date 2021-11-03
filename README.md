# `godc` a Go Desk Calculator

This is an attempt to re-implement the classic Unix program `dc`, the "Desk Calculator."

This was an exercise in becoming more familar with Go. It was
implemented from scratch to the spec of the `dc(1)` man page, not translated from the original source.

## Description

`godc` is a "Reverse Polish" calculator. That means the arguments come before the operators. `2 3 +`, not `2 + 3`. This may seem odd at first,
but it eliminates the need for an "order of operations" or parentheses. Many physical desk calculators intended for use by accountants use
"Reverse Polish" operation, because it facilitates long, running calculations.

`godc` also supports macros and registers (as does `dc`). Macros can be executed conditionally. This makes it close to a complete programming language, albeit an extremely
terse and obscure one, a little akin to assembly code. 

## Quick Start

`godc` has no dependencies apart from the standard library. But it does use the standard library's `math/big` package, so an up-to-date version of Go is required.

To build:

```
  go build
```

To run:
```
  ./godc
```

To quit, type either `q<ENTER>` or hit `CTRL+D` or `CTRL+C`.

`godc`, like all the original Unix programs that were written when Unix typed on real paper with real ink, is very terse when things are working well.
It won't automatically print the results of your calculation unless you ask it to (with `p`, `n` or `f`).

### Examples

#### Simple addition: 2 + 3

```
2 3+p<ENTER>
```

Enter `2`. Enter a space, so it won't interpret it as 23. Enter `3`. The `+` performs the operation, and the `p` shows the result.

`godc` is line-buffered, so press the `<ENTER>` key to commit your line and execute the commands.

#### Order of operations: 2 + 3 * 5

This is like one of those problems you see on Facebook, where people argue about the answer. "It's 25!" "No, it's 17!"
It's not a problem for `godc`

```
2 3 5*+p
```

Enter `2`, `3`, and `5`. The `*` multiplies the `3` and `5` and the `+` adds the `2` and `15`. The `p` prints the correct
result. (Hint: it's `17`)

If you _wanted_ (2+3) * 5, you can do that, too.

```
2 3+5*p
```

#### More complicated 15 / (2 + 3)

```
15 2 3+/p
```
Easy enough. But wait! What if you started with:

```
2 3+
```

and now you're stuck, because you forgot to put the `15` in first? Not a problem; just use the `r` operator.

```
2 3+15r/p
```

The `r` operator swaps the top two items on the stack.

#### Add up all the numbers between 1 and 25

The `d` operator duplicates the top item on the stack. Use Gauss' formula:

```
25d1+*2/p
```

That's `25`, duplicated (so you have `[25, 25]` on the stack now). `1+` adds one to the second `25`, and then `*` multiplies those together.
Then `2/` divides that by two, and `p` prints it.

#### Add up all the numbers between 1 and _n_

You wanted to do the same, but for any number, not just `25`. Or you have a lot of numbers that need the same function applied.
Write a macro. Use `[` and `]` to record strings, which can be executed as macros. Store them into registers to recall later.

```
[d1+*2/]sg
```

That stores our operations for Gauss' formula in register `g` (g for Gauss). To use it, enter your number, load register `g` with `lg`, execute it with `x`, and print the result with `p`.

```
100lgxp
```

Prints `5050` as it should.

#### Work with values less than 1.

Because `-` means "Subtract", the character to indicate the following number is negative is `_` (underscore).

```
5_5*p
```

Prints `-25`

To work with values between integers, set the precision with `k`.

```
4k2vp
```

The `v` command performs a square root. This prints `1.4142`

For other commands, see the `dc(1)` man page.

## Progress

`godc` can perform all the basic arithmetic and most macro functions of `dc`.

The following `dc` commands are not yet implemented:

- `a` Converts a number to a character, like chr(i)
- `?` Gets input from STDIN, so you can write console programs.
- `Z` Pushes the length of the top value onto the stack (digits or string length)
- `X` The number of fractional digits in the top value pushed onto the stack
- `:` Pop the top number and push it onto a regster at a specific index.
- `;` Fetch a number from a specific index in a register and push it on the stack.

`godc` also doesn't yet understand `dc`'s command-line arguments, which would
allow you to make a library of functions and populate the registers with them.
But you can do the same thing by catting your library and stdin to `godc`.