# Project Sync Tool (`pst`)


`pst` is a tool for reusing code across multiple projects on a local system, regardless of language or project type. It allows you to share specific files or folders between projects locally, without submodules, extra repositories, or package dependencies. With `pst`, you control code consistency without relying on external sources.

### Why Use This Tool?

Inspired by issues like the left-pad incident in JavaScript, `pst` offers a way to reuse code without risking external dependency failures. Instead of pulling code from an external repository or library, `pst` keeps everything local, ensuring that your projects won’t break if a dependency becomes unavailable or undergoes unexpected changes.

### Scope and Limitations
`pst` is designed to be a local utility. It operates entirely on the same device, like `cp` or `rsync`, and isn’t intended to replace package managers or manage complex dependency relationships. For example:

- **Portability**: Since `pst` tracks files with absolute paths, each user or device must configure `pst` with the same directory structure to use shared collections across systems. The configuration files are stored locally at `~/.config/project-sync-tool/`.
- **Compatibility and Testing**: `pst` only copies and updates files—it doesn’t check code compatibility between projects. Ensuring compatibility or running tests after syncing is left to the user.

---

## How It Works

### Adding Files to a Collection
To reuse a file or folder across projects, use the `share` command to add it to a named collection. This creates a copy of the specified files or folders in a central location on your system (`~/.config/project-sync-tool/collections/<collection-name>`), which becomes the source for syncing code to and from projects.

```sh
pst init <collection-name> [path/to/file/or/folder...]
```

- **Example**: Adding multiple files to a collection called `common-utils`:

  ```sh
  pst init common-utils /projectA/utils.php /projectB/helpers.php
  ```

- If no path is specified, the current directory is added to the collection.

### Requiring Files from a Collection
To update your project with the latest code from a collection, use the `require` command. This pulls changes from the central copy of each file or folder in the collection and applies them to the target path.

```sh
pst require <collection-name> [target-path]
```

- **Example**: Updating the `common-utils` collection in `projectC`:

  ```sh
  pst require common-utils /path/to/projectC
  ```

- If no target path is specified, `update` applies to the current directory.

### Syncing Collections
The `sync` command updates all collections found in the current directory or its subdirectories. Sync only applies to files that are out of sync (ignoring files that are ahead), or you can use the `--update` flag to automatically push updates from the current project to central. If both central and a project have new versions of the same file, the sync fails entirely to prevent conflicts.

```sh
pst sync [collection-name...] [--global] [--update]
```

- **Examples**:
  - Sync all collections in the current project:

    ```sh
    pst sync
    ```

  - Sync specific collections in the current project:

    ```sh
    pst sync common-utils configs
    ```

  - Sync all collections across all projects:

    ```sh
    pst sync --global
    ```

---

## Commands Overview

| Status | Command                                        | Description                                                       |
|--------|------------------------------------------------|-------------------------------------------------------------------|
|    90% | `init <name> [path(s)...]` | Add files or folders to a named collection.                       |
|    30% | `require <name>`           | Pull collection updates to a current directory or target project. |
|    90% | `push [name...]`           | Push new changes. If no collection names are provided it will scan for collections matching the current dir or target dir if provided  |
|     0% | `sync [name...] [--global] [--update]`         | Sync collections in the current directory or globally.            |
|     0% | `status [name...] [--files-only]`              | Show tracked files and folders in each collection.                |
|     0% | `add <name> [path(s)...]`                      | Add more files or folders to an existing collection.              |
|     0% | `remove <name> <path>`                         | Remove a file or folder from a collection.                        |

---

## Installation

1. **Install Go** (if not installed).
2. Clone this repository.
3. Build the tool:

   ```sh
   go build -o pst && mv pst /usr/local/bin/
   ```

