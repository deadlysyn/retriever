.PHONY: test

test:
	aws-vault exec dev -- go test -test.v -cover
