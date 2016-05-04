pwd=$(shell pwd)
arch=$(shell echo `go env GOOS`_`go env GOARCH`)
drivers_file=src/servant/server/sql_drivers.go

.PHONY : all clean driver tarball test

all:bin/servant

DRIVERS="mysql"

driver:$(drivers_file)

$(drivers_file):
	[ -e "$(drivers_file)" ] || ( echo 'package server'; \
	echo 'import (' ; \
	for d in $(DRIVERS); do \
		case "$$d" in \
		mysql) \
			GOPATH=$(pwd) go get github.com/go-sql-driver/mysql; \
			echo '_ "github.com/go-sql-driver/mysql"' ;; \
		sqlite) \
			GOPATH=$(pwd) go get github.com/mattn/go-sqlite3;  \
			echo '_ "github.com/mattn/go-sqlite3"' ;; \
		postgresql) \
			GOPATH=$(pwd) go get github.com/lib/pq;  \
			echo '_ "github.com/lib/pq"' ;; \
		esac \
	done ; \
	echo ')' ) >"$(drivers_file)"



bin/servant:$(arch)/bin/servant
	cp -r $(arch)/bin .

linux_amd64/bin/servant:$(drivers_file)
	GOOS=linux GOARCH=amd64 GOPATH=$(pwd) GOBIN=$(pwd)/linux_amd64/bin go install src/servant.go

darwin_amd64/bin/servant:$(drivers_file) 
	GOOS=darwin GOARCH=amd64 GOPATH=$(pwd) GOBIN=$(pwd)/darwin_amd64/bin go install src/servant.go


tarball:servant.tar.gz

servant.tar.gz:bin/servant
	mkdir servant
	cp -r bin conf README.md servant
	tar -czf servant.tar.gz servant
	rm -rf servant
	
test:
	GOPATH=$(pwd) go test -coverprofile=c_server.out servant/server
	GOPATH=$(pwd) go test -coverprofile=c_conf.out servant/conf

clean:
	rm -rf servant bin pkg/*/servant "$(drivers_file)" servant.tar.gz darwin_amd64 linux_amd64 c_server.out c_conf.out


