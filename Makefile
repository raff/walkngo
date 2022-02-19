LANG=c
EXT=cpp

OPTS=
SOURCES=$(wildcard tests/*.go))
OBJECTS=$(patsubst %.go,%.$(EXT),$(wildcard tests/*.go))


%.$(EXT) : %.go
	go run . -lang=$(LANG) $(OPTS) $< > $@

all: $(OBJECTS)

clean:
	-rm -rf tests/*.cpp tests/*.py a.out
