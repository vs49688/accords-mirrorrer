# Accord's Mirrorrer

A desperate attempt to mirror https://accords-library.com before it's taken down.

How it works:

1. Builds an "index" of all the entities (posts, videos, books, etc.), saved to `state.json`.
2. Scrapes the directory listing of the assets server in case anything was missed.
3. Download all the files.

## Instructions

### I just want run it

1. Go to https://github.com/accords-mirrorrer/accords-mirrorrer/releases
2. Find the latest download for your platform (this is probably `accords-mirrorrer_Windows_x86_64.zip`).
3. Create a new folder and extract `accords-mirrorrer.exe` to it. Ensure you have plenty of space.
4. Open a Command Prompt or PowerShell window
5. Run `.\accords-mirrorrer.exe archive`
6. Wait
7. If anything fails you can restart it, and it will pick up from where it left off.

### Building
```
go build ./cmd/accords-mirrorrer
```

Or download a release.

### Usage

Create a new directory, and run the `accords-mirrorrer` from there:

```
/path/to/accords-mirrorrer archive --parallelism 4 
```

### Notes

* Please be a good netizen and set `--parallelism` to a reasonable value to avoid overloading the server.
* The process may be interrupted and resumed once the index refresh is completed.
* If you already have the index, consider specifying `--dont-refresh` to avoid re-downloading it.
* On UNIX-like systems, sending `SIGUSR1` will cause it to checkpoint the current state.
* I've only tested this on Linux. Windows binaries are provided for convenience only.

### CLI Usage
```
NAME:
   accords-mirrorrer archive - archive the library

USAGE:
   accords-mirrorrer archive [command options]

OPTIONS:
   --state-file value   state file (default: "state.json")
   --parallelism value  parallelism, <1 for GOMAXPROCS (default: 0)
   --dont-refresh       don't refresh the index, only download what we've got (default: false)
   --dont-download      don't download files, only refresh the index (default: false)
   --help, -h           show help
```

## License

This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License version 2, and only
version 2 as published by the Free Software Foundation.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA  02111-1307  USA
