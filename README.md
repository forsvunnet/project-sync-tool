# Project Sync Tool (`pst`)

`pst` is a no-nonsense tool for reusing code across multiple projects on a single system, regardless of language or project type. It allows you to share specific files or folders between projects locally, without submodules, extra repositories, or network dependencies. With `pst`, you control code consistency without relying on external sources.

### Why Use This Tool?

Inspired by issues like the left-pad incident in JavaScript, `pst` offers a way to reuse code without risking external dependency failures. Instead of pulling code from an external repository or library, `pst` keeps everything local, ensuring that your projects won’t break if a dependency becomes unavailable or undergoes unexpected changes.

---

## Features

- **Local Syncing**: Share code across projects on your device—no network dependencies.
- **Checksum-Based Updates**: Detects file changes using checksums to track version consistency.
- **Named Collections**: Organize files or folders into collections for flexible syncing.
- **One-Command Sync**: Sync collections across projects with a single command, and check status across all collections.

---

## How It Works

### Adding Files to a Collection
To reuse a file or folder across projects, use the `share` command to add it to a named collection. This creates a copy of the specified files or folders in a central location on your system (`~/.config/project-sync-tool/collections/<collection-name>`), which becomes the source for syncing code to and from projects.

```sh
pst share <collection-name> [path/to/file/or/folder...]
```

- **Example**: Adding multiple files to a collection called `common-utils`:

  ```sh
  pst share common-utils /projectA/utils.php /projectB/helpers.php
  ```

- If no path is specified, the current directory is added to the collection.

### Pulling Updates from a Collection
To update your project with the latest code from a collection, use the `update` command. This pulls changes from the central copy of each file or folder in the collection and applies them to the target path.

```sh
pst update <collection-name> [target-path]
```

- **Example**: Updating the `common-utils` collection in `projectC`:

  ```sh
  pst update common-utils /path/to/projectC
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

| Command                              | Description                                                    |
|--------------------------------------|----------------------------------------------------------------|
| `share <name> [path(s)...]`          | Add files or folders to a named collection.                    |
| `update <name> [target-path]`        | Pull collection updates to a target project.                   |
| `sync [name...] [--global] [--update]`| Sync collections in the current directory or globally.         |
| `status`                             | Show tracked files and folders in each collection.             |
| `add <name> [path(s)...]`            | Add more files or folders to an existing collection.           |
| `remove <name> <path>`               | Remove a file or folder from a collection.                     |

---

## Installation

1. **Install Go** (if not installed).
2. Clone this repository.
3. Build the tool:

   ```sh
   go build -o pst
   ```

4. Move `pst` to your PATH.
