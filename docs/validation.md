# Validation

Validation belongs to explicit config structs and loader/decode contracts. Required fields, malformed values, and application-level invariants should return classified errors that are useful for startup diagnostics but safe to print.

Rules:

- missing required fields may name the field/key;
- invalid durations, numbers, or booleans may describe the expected type;
- secret values must not be included in error text;
- post-decode `Validate() error` implementations should sanitize their own messages before returning them.
