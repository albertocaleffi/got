Got - Type safe Go(lang) templating
===

Got is an [ERb](http://ruby-doc.org/stdlib-2.1.0/libdoc/erb/rdoc/ERB.html) style
templating language for Go. It works by transpiling templates into pure Go and
including them at compile time. These templates are lightweight wrappers around the
Go language itself.

## Usage

To install Got:

```sh
$ go get github.com/albertocaleffi/got/...
```

Then run `got` on a directory. Recursively traverse the directory structure and generate
Go files for all matching `.got` files.

```sh
$ got mypkg
```


## How to Write Templates

A got template lets you write text that you want to print out lets you inject actual
Go code. This means you don't need to learn a new scripting language to write got
templates—you already know Go!

### Raw Text

Any text the `got` tool encounters that is not wrapped in `<%` and `%>` tags is
considered raw text. If you have a template like this:

```
hello!
goodbye!
```

Then `got` will generate a matching `.got.go` file:

```
io.WriteString(w, "hello!\ngoodbye!")
```

Unfortunately that file won't run because we're missing a `package` line at the top.
We can fix that with _code blocks_.


### Code Blocks

A code block is a section of your template wrapped in `<%` and `%>` tags.
It is raw Go code that will be inserted into our generate `.got.go` file as-is.

For example, given this template:

```
<%
package myapp

import "io"

func Render(w io.Writer) {
%>
hello!
goodbye!
<%
}
%>
```

The `got` tool will generate:

```
package myapp

import "io"

func Render(w io.Writer) {
	io.WriteString(w, "hello!\ngoodbye!")
}
```

_Note the `io` package must be imported to your template._
_You'll need to import any other packages you use._

## Caveats

Unlike other runtime-based templating languages, Got does not support ad hoc templates.
All Got templates must be generated into Go code before compile time.

Got does not attempt to provide any security around the templates. Just like regular Go
code, the security model is up to you. The text rendered in Got templates must be
escaped to prevent XSS (Cross-Site Scripting) and other web vulnerabilities.

## Credits

Got is based on the [Ego](https://github.com/benbjohnson/ego) templating package by
Ben Johnson.
