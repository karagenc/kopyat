# Kopyat

Kopyat is my Swiss Army knife for backup and sync related tasks. The name Kopyat comes from 2 words: kopya (Turkish for "copy") and cat -> copycat.

Functionalities:
- Serve as a wrapper for backup programs (for now, only supported backup program is restic), optionally providing ifile support.
- Generate ifile (`.stignore`) for syncthing directories.

## Ifile

Ifile is a type of file generated from `.gitignore` and `.kopyatignore` (same format as `.gitignore`) files found inside the directory tree.  

"I" of the ifile stands for both ignore and include. For restic backups, it generates an simple straightforward include file, and for syncthing, it generates an ignore file (`.stignore`).

## Build

```shell
git clone https://github.com/karagenc/kopyat
cd kopyat
make
```

## Config

**Note:** Paths in config are relative to the config file. Exception to this are backup base path and backup paths. They must be absolute in order for ifile to be generated without problems.

### Environment variables

- `KOPYAT_CONFIG`
- `KOPYAT_STATE_DIR`
- `KOPYAT_SCRIPT`
  - Set to `1` only inside Go hooks.

### Hooks (Scripts)

Kopyat can run hooks before or after backup and ifile generation operations. Hooks are simply Go source files with an `init` function. Save the code snippet below as `hello.go`, and run it with: `kopyat run-script hello.go`

```go
package x
//      ^ package name can be anything, including main.

import "fmt"

func init() {
  fmt.Println("Hello from hook script")
}
```

To access context-specific information, import `kopyat` and:

```go
package x
//      ^ package name can be anything, including main.

import "github.com/karagenc/kopyat"

func init() {
  // Get Kopyat context.
  // Only for backup and ifile generation hooks.
  c := kopyat.GetContext()
  // Access context
  // ...
}
```

