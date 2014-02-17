/*
 * GoBoat, the Boat that goes.
 */

package main

import (
  "code.google.com/p/gcfg"
)

// load and set defaults

func load_config(file string) Config {
  var cfg Config
  err := gcfg.ReadFileInto(&cfg, file)
  fail(err)

  //if strings.ToLower(cfg.Logger.Driver) != "sqlite3" { }

  // use some sane defaults
  for net,v := range cfg.Network {
    // must have a server to connect to
    if v.Host == "" {
      failStrf("load_config: no host given in network %s", net)
    }

    // must have a nickname
    if v.Nick == "" {
      if cfg.General.Nick == "" {
        failStrf("load_config: no nickname given in network %s", net)
      } else { v.Nick = cfg.General.Nick }
    }

    // defaults
    if v.Port     == 0  { v.Port     = 6667 }
    if v.Username == "" {
      if cfg.General.Username == "" {
        v.Username = "GoBoat"
      } else {
        v.Username = cfg.General.Username
      }
    }
  }

  return cfg
}
