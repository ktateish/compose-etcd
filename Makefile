.PHONY: all clean

all: config.yaml gen tmpl/compose-etcd.yaml
	./gen < config.yaml

clean:
	rm -rf gen compose/*

gen: gen.go
	go build gen.go
