The following are required to run:

    go get "code.google.com/p/gcfg"
    go get "github.com/thoj/go-ircevent"
    go get "github.com/mattn/go-sqlite3"

Then provide a configuration file as described by `config.ini.example`
to the executable.

    go run nullboat.go -config=<config.ini>

This option can be omitted if `config.ini` is located in the same
directory as `nullboat.go`.

Enjoy.
