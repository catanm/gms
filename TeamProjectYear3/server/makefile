run:
	go fmt main.go
	go fmt packages/gms/endpoints/endpoints.go
	go fmt packages/gms/models/models.go
	go fmt packages/gms/log/log.go
	go fmt packages/gms/utils/utils.go
	cp -r ./packages/* $(GOPATH)/src/
	go run main.go
	

runprocessing:
	go fmt processing.go
	go fmt packages/gms/endpoints/endpoints.go
	go fmt packages/gms/models/models.go
	go fmt packages/gms/log/log.go
	go fmt packages/gms/utils/utils.go
	cp -r ./packages/* $(GOPATH)/src/
	go run processing.go

runsprocessing:
	go fmt processing.go
	go fmt packages/gms/endpoints/endpoints.go
	go fmt packages/gms/models/models.go
	go fmt packages/gms/log/log.go
	go fmt packages/gms/utils/utils.go
	cp -r ./packages/* $(GOPATH)/src/
	go run processing.go >processingLog.txt &

runserver:
	go fmt main.go
	go fmt packages/gms/endpoints/endpoints.go
	go fmt packages/gms/models/models.go
	go fmt packages/gms/log/log.go
	go fmt packages/gms/utils/utils.go
	cp -r ./packages/* $(GOPATH)/src/
	go run main.go >logs.txt &

findprocess:
	ps axf | grep "go run main.go"

update:
	git pull
	cp -r ./packages/* $(GOPATH)/src/

setup:
	# go get -v -u gopkg.in/mgo.v2 # shoudl be the same as labix's mgo, goweb uses the labix one, so this way there is no duplicate packages
	go get -v -u labix.org/v2/mgo
	go get -v -u github.com/stretchr/goweb
	go get -v -u github.com/stretchr/gomniauth
	go get -v -u github.com/LarryBattle/nonce-golang
	go get -v -u github.com/rwcarlsen/goexif/exif
	go get -v -u golang.org/x/crypto/bcrypt
	go get -v -u github.com/opennota/phash

format:
	go fmt main.go
	go fmt packages/gms/endpoints/endpoints.go
	go fmt packages/gms/models/models.go
	go fmt packages/gms/log/log.go
	go fmt packages/gms/utils/utils.go

clean:
	rm -rf $(GOPATH)/src/gms

cleanall:
	rm -rf $(GOPATH)/src/labix.org/v2/mgo
	rm -rf $(GOPATH)/src/github.com/stretchr/goweb
	rm -rf $(GOPATH)/src/github.com/stretchr/gomniauth
	rm -rf $(GOPATH)/src/github.com/LarryBattle/nonce-golang
	rm -rf $(GOPATH)/src/github.com/rwcarlsen/goexif/exif
	rm -rf $(GOPATH)/src/golang.org/x/crypto/bcrypt

cleantotal:
	rm -rf $(GOPATH)/src
	rm -rf $(GOPATH)/pkg
