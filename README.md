# bwtui

A simple terminal user interface for [Bitwarden](https://bitwarden.com/)

## Prerequisites

bwtui currently only works on Linux. You need the following tools installed to make use of the full feature range:

* xsel: for copying username/password to the X11 clipboard
* [bw](https://github.com/bitwarden/cli): the Bitwarden CLI

## Usage

After starting up, bwtui will fetch all items stored in your Vault and list them.

| Key | Action                                                     |
|-----|------------------------------------------------------------|
| `u` | copy the username of the highlighted item to the clipboard |
| `p` | copy the password of the highlighted item to the clipboard |
| `/` | filter items by name                                       |
| `q` | quit bwtui                                                 |
| `r` | in the details dialog this reveals the password            |

The UI also supports the VIM movement controls `j` (one item down) and `k` (one item up).

## License

GNU General Public License v3.0 or later
