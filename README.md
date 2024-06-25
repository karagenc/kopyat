# kopyaship

The name kopyaship comes from 2 words: kopya (Turkish for "copy") and ship -> copyship. It is pronounced as copiaship.

## Config

- Paths in config are relative to the config file. Exception to this are backup base path and backup paths. They must be absolute in order for ifile to be generated without problems.

### Environment variables

- `KOPYASHIP_CONFIG`
- `KOPYASHIP_STATE_DIR`
- `KOPYASHIP_SCRIPT`
  - Set to `1` only inside Go hooks.

### Hooks (Scripts)

```go
package x
//      ^ package name can be anything, including main.

import "github.com/tomruk/kopyaship"

func init() {

}
```

