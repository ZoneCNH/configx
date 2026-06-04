# Sanitized results

`LoadResult.Sanitize()` is the safe reporting path for loaded configuration. It returns a `SanitizedResult` with effective values and trace metadata suitable for diagnostics and release evidence.

Sanitized output may contain source names, kinds, paths selected by the caller, key names, hashes, check statuses, and redacted markers. It must not contain raw values for keys classified as secret.

Any new report or JSON output type must include a negative test proving representative secret material is absent.
