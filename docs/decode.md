# Decode

`Decode(result, &target)` converts a `LoadResult` into exported struct fields. It supports `config`, `default`, `required`, and `config:"-"` tags plus nested structs.

Supported field shapes include string, bool, signed and unsigned integers, floats, `time.Duration`, `SecretString`, and types implementing `encoding.TextUnmarshaler`. If the populated target implements `Validate() error`, validation runs after assignment.

Decode errors must identify fields and conversion problems without embedding raw secret values.
