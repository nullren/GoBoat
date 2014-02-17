/*
 * GoBoat, the Boat that goes.
 */

package main

import (
  "log"
  "fmt"
)

func fail(err error) {
  if err != nil {
    log.Fatal(err)
  }
}

func failStrf(msg string, a ...interface{}) {
  fail(fmt.Errorf(msg, a))
}
