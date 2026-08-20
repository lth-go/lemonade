Lemonade
========

remote...lemote...lemode......Lemonade!!! :lemon: :lemon:

Lemonade is a remote utility tool.
(copy, paste and open browser) over HTTP.


Installation
------------

```sh
go get -d github.com/lemonade-command/lemonade
cd $GOPATH/src/github.com/lemonade-command/lemonade/
make install
```

Example of use
----------------

![Example](http://f.st-hatena.com/images/fotolife/P/Pocke/20150823/20150823173041.gif)

For example, you use a Linux as a virtual machine on Windows host.
You connect to Linux by SSH client(e.g. PuTTY).
When you want to copy text of a file on Linux to Windows, what do you do?
One solution is doing `cat file.txt` and drag displayed text.
But this answer is NOT elegant! Because your hand leaves from the keyboard to use the mouse.

Another solution is using the Lemonade.
You input `cat file.txt | lemonade copy`. Then, lemonade copies text of the file to clipboard of the Windows!

In addition to the above, lemonade supports pasting and opening URL.


Usage
--------

```sh
Usage: lemonade [options]... SUB_COMMAND [arg]
Sub Commands:
  copy [text]                 Copy text.
  paste                       Paste text.
  server                      Start lemonade server.

Options:
  --port=2489                 TCP port number
  --line-ending               Convert Line Ending(CR/CRLF)
  --allow="0.0.0.0/0,::/0"    Allow IP Range                [Server only]
  --host="localhost"          Destination hostname          [Client only]
  --help                      Show this message
```


### On server (in the above, Windows)

```sh
$ lemonade server
```


### Client (in the above, Linux)


```sh
# You want to copy a text
$ cat file.txt | lemonade copy

# You want to paste a text from the clipboard of Windows
$ lemonade paste
```


Configuration
---------------

You can override command line options by configuration file.
There is configuration file at `~/.config/lemonade.toml`.

### Server

```toml
port = 1234
allow = '192.168.0.0/24'
line-ending = 'crlf'
```

- `port` is a listening port of TCP.
- `allow` is a comma separated list of a allowed IP address(with CIDR block).


### Client

```toml
port = 1234
host = '192.168.x.x'
line-ending = 'crlf'
```

- `port` is a port of server.
- `host` is a hostname of server.


Advanced Usage
-----------------

### line-ending

Default: "" (NONE)

This options works with `copy` and `paste` command only.

If this option is `lf` or `crlf`, lemonade converts the line ending of text to the specified.
