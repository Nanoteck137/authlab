run:
	air

clean:
	rm -rf work
	rm -rf tmp
	rm -f result

test-build:
	nix build --no-link .#backend
	nix build --no-link .#frontend

publish: test-build
	publish-version

.PHONY: run clean publish
