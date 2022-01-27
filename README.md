Performance tests
=================

This is an open source benchmark to test the performance of ObjectBox and other databases that offer persistence of Go structs (e.g. GORM, Storm/bbolt).

Tests include:

* CRUD (create, read, update, delete) operations using batches of structs
* Lookup by IDs
* Query by prefix

How to run
----------

For each database to test there's a command line executable represented by a main.go file in those directories:

* objectbox
* gorm
* bolt-storm

The following examples refer to the objectbox directory, but you can do the same for others.  

To get good numbers, close all programs before running, build & run outside of IDE:

```shell script
cd objectbox
go build 
./objectbox
```

or you can use `go run ./objectbox` which does yield about the same results

Parameters
----------

A typical invocation looks like this (e.g. 3 runs with 100K objects each run):

```
./objectbox -count 100000 -runs 3
```

You can specify some parameters, see `./objectbox -h`:
```
Usage of ./objectbox:
  -count int
    	number of objects (default 10000)
  -db string
    	database directory (default "testdata")
  -runs int
    	number of times the tests should be executed (default 10)
```

Dev notes
---------
To regenerate ObjectBox entity bindings
```shell script
mv objectbox/obx/objectbox-model.* internal/models/
go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen internal/models/entity.go
mv internal/models/entity.obx.go objectbox/obx/
mv internal/models/objectbox-model.* objectbox/obx/
for f in objectbox/obx/*.go; do sed -i 's/package models/package obx/g' "$f"; done
for f in objectbox/obx/*.obx.go; do sed -i '7 a 	. "github.com/objectbox/objectbox-go-performance/internal/models"' "$f"; done
```