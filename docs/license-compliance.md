# License Compliance

GitHub CLI complies with the software licenses of its dependencies. This document explains how license compliance is maintained.

## Overview

When a dependency is added or updated, the license information needs to be updated. We use the [`google/go-licenses`](https://github.com/google/go-licenses) tool to:

1. Generate markdown documentation listing all Go dependencies and their licenses
2. Copy license files for dependencies that require redistribution

## License Files

The following files contain license information:

- `third-party-licenses.darwin.md` - License information for macOS dependencies
- `third-party-licenses.linux.md` - License information for Linux dependencies
- `third-party-licenses.windows.md` - License information for Windows dependencies
- `third-party/` - Directory containing source code and license files that require redistribution

## Updating License Information

When dependencies change, you need to update the license information:

1. Update license information for all platforms:

   ```shell
   script/licenses
   ```

2. Commit the changes:

   ```shell
   git add third-party-licenses.*.md third-party/
   git commit -m "Update third-party license information"
   ```

## Checking License Compliance

The CI workflow checks if license information is up to date. To check locally:

```sh
script/licenses-check
```

If the check fails, follow the instructions to update the license information.

## How It Works

- `script/licenses` - Script to generate license information for all platforms or a specific platform
- `script/licenses-check` - Script to check if license information is up to date
- `.github/workflows/lint.yml` - CI workflow that includes license compliance checking
