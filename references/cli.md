# Go CLI conventions (cobra)

## Layout

```text
cmd/<binary>/main.go   build Factory, root command, exit code
pkg/cmdutil/           Factory, scope resolution, server-error mapping, completions
pkg/cmd/<group>/       one file per verb: get.go, list.go, delete.go (+ _test.go)
pkg/iostreams/         In/Out/ErrOut, JSON flag, TTY detection, color scheme, Test()
pkg/render/            Error, JSONSuccess, JSONError, EmptyList
pkg/clierr/            CLIError type and stable error codes
pkg/clients/           generated API clients (oapi-codegen ClientWithResponses)
pkg/config/            on-disk config, instances, tokens
pkg/prompter/          interactive prompts
```

Leaf packages (`auth`, `clients`, `config`) never import `cmdutil`. They take concrete dependencies as parameters.

## Factory

`cmdutil.Factory` is the only source of shared dependencies:

```go
type Factory struct {
	Config       func() (*config.Config, error)
	IOStreams    *iostreams.IOStreams
	Prompter     prompter.Prompter
	HTTPClient   func() *http.Client
	AgentManager func(ctx context.Context) (*amsvc.ClientWithResponses, error)
	Token        func(ctx context.Context) (string, error)
}
```

Lazy dependencies are funcs so tests can replace them and token refresh happens once, in one place. Commands never call `iostreams.System()`, `config.Load()`, or build clients themselves. If the Factory lacks something, add it there.

## Command shape

```go
type GetOptions struct {
	IO           *iostreams.IOStreams
	Client       func(context.Context) (*amsvc.ClientWithResponses, error)
	ResolveScope func(*cobra.Command, bool, bool) (string, string, error)
	MakeScope    func(org, proj, name string) render.Scope

	Org, Proj string
	Scope     render.Scope
	Name      string
}

func NewGetCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &GetOptions{
		IO:           f.IOStreams,
		Client:       f.AgentManager,     // direct reference, never a forwarding closure
		ResolveScope: f.ResolveOrgProject,
		MakeScope:    f.AgentScope,
	}
	cmd := &cobra.Command{
		Use:   "get [name]",
		Short: "Show details of a widget",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			org, proj, err := opts.ResolveScope(cmd, true, true)
			if err != nil {
				return render.Error(opts.IO, opts.MakeScope(org, proj, ""), err)
			}
			opts.Org, opts.Proj = org, proj
			opts.Scope = opts.MakeScope(org, proj, args[0])
			opts.Name = args[0]
			return runGet(cmd.Context(), opts)
		},
	}
	return cmd
}

func runGet(ctx context.Context, o *GetOptions) error { ... }
```

- `RunE` is the only code that touches `*cobra.Command`. It resolves flags and scope, fills `opts`, and calls `runX(ctx, opts)`.
- `runX`, validators, and output builders never take cobra types.
- Options fields from the Factory are direct method values. If the signature doesn't match, change the field type or add a Factory helper. Closures are fine only where cobra demands a function (`ValidArgsFunction`, `RegisterFlagCompletionFunc`).
- Validate path-like arguments (`ValidatePathParam`) before any network call.

## Output and errors

- All output through `o.IO.Out` / `o.IO.ErrOut`. Never `fmt.Println` or `os.Stdout`.
- Every error path is `return render.Error(o.IO, o.Scope, err)`. Never a naked `return err`.
- Transport failures become `clierr.Newf(clierr.Transport, "%v", err)`. Non-2xx responses go through `cmdutil.ErrorFromServer(resp.HTTPResponse, firstNonNil(resp.JSON404, resp.JSON500))`.
- `--json` mode: `render.JSONSuccess(o.IO, o.Scope, data)` writes `{"scope":..., "data":...}`. `render.JSONError` writes the error envelope. Both are wire contracts for scripts.
- `clierr.CLIError{Status, Code, Message, Reason *string, AdditionalData map[string]any}` is both a Go error and the JSON error body. `AdditionalData` is always non-nil so it marshals to `{}`. Codes are stable constants. Add, never rename.
- Text mode: aligned `label:  value` lines, `cs.Bold` for the name, `cs.Gray` for timestamps, `-` for absent values. Timestamps in RFC 3339.
- A command that already printed its own output returns `render.Rendered(err)` so the root exits non-zero without printing twice.

## Testing

Build `*XxxOptions` by hand and call `runX(context.Background(), opts)`. No cobra tree.

```go
func TestGet_RejectsEmptyName(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	ios.JSON = true
	err := runGet(context.Background(), &GetOptions{IO: ios, Client: unreachableClient, Scope: baseScope(), Name: ""})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	env := decodeEnvelope(t, out.String())
	if env["error"].(map[string]any)["code"] != clierr.InvalidFlag { ... }
}
```

- `iostreams.Test()` returns streams backed by buffers.
- A client func that fails the test if called (`unreachableClient`) proves validation short-circuits.
- An `httptest.Server` stands in for the API when the happy path matters.
- Assert on the JSON envelope's `code`, not on message text.

## Doc comments

Exported identifiers always have doc comments. Shared packages (`cmdutil`, `render`, `clierr`, `clients`) document unexported helpers too. Comments say why it exists and when to use it, not what the signature already says.
