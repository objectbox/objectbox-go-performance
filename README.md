Performance tests
=================

To get good numbers, close all programs before running, build & run outside of IDE:

```shell script
cd objectbox
go build 
./objectbox
```

or you can use `go run ./objectbox` which does yield about the same results

Parameters
----------
You can specify some parameters, see `./objectbox -h`:
```
Usage of ./objectbox:
  -count int
    	number of objects (default 1000)
  -db string
    	database directory (default "testdata")
  -runs int
    	number of times the tests should be executed (default 10)
```

Contributing
------------
To regenerate ObjectBox entity bindings
```shell script
go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen -source internal/models/entity.go
mv internal/models/entity.obx.go objectbox/obx/
mv internal/models/objectbox-model.* objectbox/obx/
for f in objectbox/obx/*.go; do sed -i 's/package models/package obx/g' "$f"; done
for f in objectbox/obx/*.obx.go; do sed -i '7 a 	. "github.com/objectbox/go-benchmarks/internal/models"' "$f"; done
```
and add a missing import `. "github.com/objectbox/go-benchmarks/internal/models"` to `entity.obx.go`