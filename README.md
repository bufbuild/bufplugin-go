# bufplugin-go

[![Build](https://github.com/bufbuild/bufplugin-go/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/bufbuild/bufplugin-go/actions/workflows/ci.yaml)
[![Report Card](https://goreportcard.com/badge/buf.build/go/bufplugin)](https://goreportcard.com/report/buf.build/go/bufplugin)
[![GoDoc](https://pkg.go.dev/badge/buf.build/go/bufplugin.svg)](https://pkg.go.dev/buf.build/go/bufplugin)
[![Slack](https://img.shields.io/badge/slack-buf-%23e01563)](https://buf.build/links/slack)

This is the Go SDK for the [Bufplugin](https://github.com/bufbuild/bufplugin) framework.
`bufplugin-go` currently provides the [check](https://pkg.go.dev/buf.build/go/bufplugin/check),
[checkutil](https://pkg.go.dev/buf.build/go/bufplugin/check/checkutil), and
[checktest](https://pkg.go.dev/buf.build/go/bufplugin/check/checktest) packages to make it simple to
author _and_ test custom lint and breaking change plugins. It wraps the `bufplugin` API with
[pluginrpc-go](https://github.com/pluginrpc/pluginrpc-go) in easy-to-use interfaces and concepts
that organize around the standard
[protoreflect](https://pkg.go.dev/google.golang.org/protobuf@v1.34.2/reflect/protoreflect) API that powers
most of the Go Protobuf ecosystem. `bufplugin-go` is also the framework that the Buf team uses to
author all of the builtin lint and breaking change rules within the
[Buf CLI](https://github.com/bufbuild/buf) - we've made sure that `bufplugin-go` is powerful enough
to represent the most complex lint and breaking change rules while keeping it as simple as possible
for you to use. If you want to author a lint or breaking change plugin today, you should use
`bufplugin-go`.

## Use

A plugin is just a binary on your system that implements the
[Bufplugin API](https://buf.build/bufbuild/bufplugin). Once you've installed a plugin, simply add a
reference to it and its rules within your `buf.yaml`. For example, if you've installed the
[buf-plugin-timestamp-suffix](check/internal/example/cmd/buf-plugin-timestamp-suffix) example plugin
on your `$PATH`:

```yaml
version: v2
lint:
  use:
    - STANDARD # omit if you do not want to use the rules builtin to buf
    - TIMESTAMP_SUFFIX
plugins:
  - plugin: buf-plugin-timestamp-suffix
    options:
      timestamp_suffix: _timestamp # set to the suffix you'd like to enforce
```

All configuration that can be used for builtin rules can be used for rules exposed by plugins; the
`use, except, ignore, ignore_only` keys work just as you'd expect.

Plugins can be named whatever you'd like them to be, however we'd recommend following the convention
of prefixing your binary names with `buf-plugin-` for clarity.

## Examples

In this case, examples are worth a thousand words, and we recommend you read the examples in
[check/internal/example/cmd](check/internal/example/cmd) to get started:

- [buf-plugin-timestamp-suffix](check/internal/example/cmd/buf-plugin-timestamp-suffix): A simple
  plugin that implements a single lint rule, `TIMESTAMP_SUFFIX`, that checks that all
  `google.protobuf.Timestamp` fields have a consistent suffix for their field name. This suffix is
  configurable via plugin options.
- [buf-plugin-field-lower-snake-case](check/internal/example/cmd/buf-plugin-field-lower-snake-case):
  A simple plugin that implements a single lint rule, `PLUGIN_FIELD_LOWER_SNAKE_CASE`, that checks
  that all field names are `lower_snake_case`.
- [buf-plugin-field-option-safe-for-ml](check/internal/example/cmd/buf-plugin-field-option-safe-for-ml):
  Likely the most interesting of the examples. A plugin that implements a lint rule
  `FIELD_OPTION_SAFE_FOR_ML_SET` and a breaking change rule `FIELD_OPTION_SAFE_FOR_ML_STAYS_TRUE`,
  both belonging to the `FIELD_OPTION_SAFE_FOR_ML` category. This enforces properties around an
  example custom option `acme.option.v1.safe_for_ml`, meant to denote whether or not a field is safe
  to use in ML models. An organization may want to say that all fields must be explicitly marked as
  safe or unsafe across all of their schemas, and no field changes from safe to unsafe. This plugin
  would enforce this organization-side. The example shows off implementing multiple rules,
  categorizing them, and taking custom option values into account.
- [buf-plugin-syntax-specified](check/internal/example/cmd/buf-plugin-syntax-specified): A simple
  plugin that implements a single lint rule, `PLUGIN_SYNTAX_SPECIFIED`, that checks that all files
  have an explicit `syntax` declaration. This demonstrates using additional metadata present in the
  `bufplugin` API beyond what a `FileDescriptorProto` provides.

All of these examples have a `main.go` plugin implementation, and a `main_test.go` test file that
uses the `checktest` package to test the plugin behavior. The `checktest` package uses
[protocompile](https://github.com/bufbuild/protocompile) to compile test `.proto` files on the fly,
run them against your rules, and compare the resulting annotations against an expectation.

Here's a short example of a plugin implementation - this is all it takes:

```go
package main

import (
	"context"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func main() {
	check.Main(
		&check.Spec{
			Rules: []*check.RuleSpec{
				{
					ID:      "PLUGIN_FIELD_LOWER_SNAKE_CASE",
					Default: true,
					Purpose: "Checks that all field names are lower_snake_case.",
					Type:    check.RuleTypeLint,
					Handler: checkutil.NewFieldRuleHandler(checkFieldLowerSnakeCase, checkutil.WithoutImports()),
				},
			},
		},
	)
}

func checkFieldLowerSnakeCase(
	_ context.Context,
	responseWriter check.ResponseWriter,
	_ check.Request,
	fieldDescriptor protoreflect.FieldDescriptor,
) error {
	fieldName := string(fieldDescriptor.Name())
	fieldNameToLowerSnakeCase := toLowerSnakeCase(fieldName)
	if fieldName != fieldNameToLowerSnakeCase {
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"Field name %q should be lower_snake_case, such as %q.",
				fieldName,
				fieldNameToLowerSnakeCase,
			),
			check.WithDescriptor(fieldDescriptor),
		)
	}
	return nil
}

func toLowerSnakeCase(fieldName string) string {
	// The actual logic for toLowerSnakeCase would go here.
	return "TODO"
}
```

## Status: Beta

Bufplugin is currently in beta, and may change as we work with early adopters. We're intending to
ship a stable v1.0 by the end of 2024. However, we believe the API is near its final shape.

## Legal

Offered under the [Apache 2 license](https://github.com/bufbuild/bufplugin-go/blob/main/LICENSE).
