go fmt main.go
go fmt packages/gms/endpoints/endpoints.go
go fmt packages/gms/models/models.go
go fmt packages/gms/log/log.go
go fmt packages/gms/utils/utils.go
cp -r ./packages/* %GOPATH%/src/
go run main.go