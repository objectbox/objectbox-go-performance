module github.com/objectbox/go-benchmarks

go 1.12

require (
	cloud.google.com/go v0.41.0 // indirect
	github.com/google/flatbuffers v1.11.0
	github.com/jinzhu/gorm v1.9.10
	github.com/objectbox/objectbox-go v1.0.0
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
)

replace github.com/objectbox/objectbox-go v1.0.0 => ../objectbox-go
