# Secret handling

Secrets are values that should not be printable by default. `configx` supports explicit secret marking through `NewSecretMapSource`, secret decode targets, and key-name heuristics such as password, token, access key, and secret key.

Use `SecretString` for secret fields. Its string, text, and JSON output is redacted. Access to raw material must stay inside trusted application code and must not be used in examples, logs, tests, reports, or release evidence.
