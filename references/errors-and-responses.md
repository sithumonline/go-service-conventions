# Errors and responses

## Three kinds of error

1. **Sentinels** in `utils/errors.go`, one `var (...)` block grouped by comment (`// Resource not found errors`, conflict, precondition). Message is lower-case, no punctuation: `errors.New("widget not found")`. A sentinel whose status mapping is not obvious carries a comment ending `Maps to 409.`
2. **`ValidationError`** for client mistakes that need a user-facing message and a technical reason:

```go
type ValidationError struct {
	Message  string // shown in the UI: short, non-technical, actionable
	Reason   string // for developers: field names, rules
	sentinel error  // optional, so errors.Is(err, ErrInvalidInput) still matches
}
func (e *ValidationError) Error() string { return e.Reason }
func (e *ValidationError) Unwrap() error { return e.sentinel }
```

Constructors: `NewValidationError(msg, reason)`, `NewValidationErrorf(msg, reasonFmt, args...)`, `NewInvalidInputError(msg, reason)`. Detect with `IsValidationError(err) *ValidationError` (wraps `errors.As`).

3. **Wrapped errors** for everything else. Service: `fmt.Errorf("failed to list widgets: %w", err)`. Outbound client: `fmt.Errorf("widgetsvc.GetWidget: %w", err)` with a `pkg.Func:` prefix.

## Error codes

`utils/error_codes.go` holds string constants in SCREAMING_SNAKE_CASE. Generic codes first (`VALIDATION_ERROR`, `NOT_FOUND`, `CONFLICT`, `BAD_REQUEST`, `UNAUTHORIZED`, `FORBIDDEN`, `INTERNAL_ERROR`, `SERVICE_UNAVAILABLE`), then resource-specific ones (`WIDGET_NOT_FOUND`, `WIDGET_ALREADY_EXISTS`). The frontend branches on these, so they never change once shipped.

## Response writers

```go
func WriteSuccessResponse[T any](w http.ResponseWriter, statusCode int, data T)
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string)          // code derived from status
func WriteErrorResponseWithReason(w http.ResponseWriter, statusCode int, message, reason, code string)
func WriteValidationErrorResponse(w http.ResponseWriter, err error)                    // 400 VALIDATION_ERROR
```

Each sets `Content-Type: application/json`, writes the status, and encodes. 204 writes no body. Encoding errors after the header is written are ignored with `_ =` and a comment.

Error body:

```json
{"message": "Widget not found", "code": "WIDGET_NOT_FOUND", "reason": "widget not found"}
```

`message` is for people, `code` is for programs, `reason` is the technical detail (usually `err.Error()` of the sentinel, never an internal wrapped chain with hostnames or SQL).

## The controller's error switch

One `handleCommonErrors(w, err, fallbackMsg)` per controller package:

```go
func handleCommonErrors(w http.ResponseWriter, err error, fallbackMsg string) {
	switch {
	// Keep FIRST: a ValidationError may also match a sentinel below, which
	// would replace its specific Message with a generic one.
	case utils.IsValidationError(err) != nil:
		utils.WriteValidationErrorResponse(w, err)

	// Not found
	case errors.Is(err, utils.ErrWidgetNotFound):
		utils.WriteErrorResponseWithReason(w, http.StatusNotFound, "Widget not found", err.Error(), utils.ErrCodeWidgetNotFound)

	// Conflict
	case errors.Is(err, utils.ErrWidgetAlreadyExists):
		utils.WriteErrorResponseWithReason(w, http.StatusConflict, "Widget already exists", err.Error(), utils.ErrCodeWidgetAlreadyExists)

	default:
		utils.WriteErrorResponse(w, http.StatusInternalServerError, fallbackMsg)
	}
}
```

- The default case returns the caller's `fallbackMsg` ("Failed to get widget"), never `err.Error()`. Details go to the log line the handler wrote just before.
- A handler that wants a different status for one sentinel checks it with `errors.Is` before calling `handleCommonErrors`, with a comment saying why.
- Panics are caught by recovery middleware, logged with stack and correlation ID, and answered with 500 `"internal server error"`.

## Logging errors

- `log.Error("Failed to get widget", "error", err)` in the controller, once. Services return errors and do not log them too.
- Client mistakes (bad label, bad version tag) log at Debug with the handler name: `log.Debug("UpdateWidget: invalid labels", "error", err)`.
- Upstream partial failures in a fan-out log as Warn and the request continues.
- Never log a secret or a request body.
