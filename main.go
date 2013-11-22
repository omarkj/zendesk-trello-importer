package main

import (
  "flag"
  "log"
)

func main() {
  var importer, webserver bool
  flag.BoolVar(&importer, "importer", false, "run as an importer")
  flag.BoolVar(&webserver, "webserver", false, "run as a webserver")
  flag.Parse()

  if !(importer || webserver) || (importer && webserver) {
    log.Fatal("must specify either -importer, -webserver")
  }

  if importer {
    runAsImporter()
  }

  if webserver {
    runAsWebserver()
  }
}

