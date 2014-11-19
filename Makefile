tinycat:
	go build -tags leveldb cmd/tinycat/tinycat.go

clean:
	rm -f tinycat

clean-all: clean
	rm -rf tc.bleve
