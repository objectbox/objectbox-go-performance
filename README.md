Performance tests
=================

To get good numbers, close all programs before running, build & run outside of IDE:

```bash
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
