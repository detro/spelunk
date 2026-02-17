# JSONPath Secret Modifier (`jp`)

This modifier allows you to extract specific values from a secret that contains JSON data. It uses [JSONPath](https://goessner.net/articles/JsonPath/) syntax to query the JSON structure.

## Status

**Built-in**: This modifier is included and enabled by default in `spelunk`.

## Usage

To use the JSONPath modifier, append `?jp=<expression>` to your secret coordinates URI.

### Syntax

```
<type>://<location>?jp=<jsonpath_expression>
```

- **Modifier Key**: `jp`
- **Value**: A valid JSONPath expression (e.g., `$.users[0].name`).

### Example

Suppose you have a secret stored in Kubernetes at `my-namespace/db-config/connection` with the following JSON content:

```json
{
  "host": "db.example.com",
  "port": 5432,
  "users": [
    { "username": "admin", "role": "read-write" },
    { "username": "viewer", "role": "read-only" }
  ]
}
```

To extract just the **host**:
```
k8s://my-namespace/db-config/connection?jp=$.host
```
**Result**: `db.example.com`

To extract the **username of the first user**:
```
k8s://my-namespace/db-config/connection?jp=$.users[0].username
```
**Result**: `admin`

## Behavior

1.  **Parsing**: The modifier first attempts to parse the retrieved secret value as JSON. If the secret is not valid JSON, it returns an error.
2.  **Extraction**: It applies the provided JSONPath expression.
3.  **Result Handling**:
    -   **Strings**: Returned as-is.
    -   **Numbers**: Converted to string, with trailing zeros removed (e.g., `1.500` becomes `1.5`).
    -   **Booleans**: Converted to string (`"true"` or `"false"`).
    -   **Lists/Arrays**: If the JSONPath expression matches multiple elements, **only the first element is returned**.
    -   **Objects/Complex Types**: Marshaled back into a JSON string.
    -   **Null**: Returns an error indicating the result is null.

## Implementation Details

This modifier uses the [github.com/oliveagle/jsonpath](https://github.com/oliveagle/jsonpath) library, which implements the [RFC-9535](https://www.rfc-editor.org/rfc/rfc9535) standard for JSONPath.
