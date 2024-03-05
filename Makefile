WORKDIR = .

default: parser

parser:
	@go build -o $(WORKDIR)/bin/parser $(WORKDIR)/*.go >/dev/null;