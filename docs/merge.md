# Merge precedence

`Loader` merges sources in the order supplied by the caller. Later source values win over earlier values, and overwritten values remain visible only as source/provenance metadata.

The merge contract is:

1. caller orders sources explicitly with `AddSource`;
2. each source reports loaded key names and sanitized failures;
3. effective values are traced to the winning source;
4. sanitized reports never include raw secret material.

Future duplicate-detection modes may be added, but they must remain explicit loader options and must preserve existing redaction guarantees.
