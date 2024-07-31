# kopyat

The name kopyat comes from 2 words: kopya (Turkish for "copy") and cat -> copycat.

## Config

- Paths in config are relative to the config file. Exception to this are backup base path and backup paths. They must be absolute in order for ifile to be generated without problems.

### Environment variables

- `KOPYAT_CONFIG`
- `KOPYAT_STATE_DIR`
- `KOPYAT_SCRIPT`
  - Set to `1` only inside Go hooks.

### Hooks (Scripts)

```go
package x
//      ^ package name can be anything, including main.

import "github.com/karagenc/kopyat"

func init() {

}
```

