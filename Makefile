default:
	make clean
	make unittest
	make dist
	make release

unittest:
	@echo "Running tests..."
	. venv/bin/activate; python -m unittest discover -s tests -p '*_test.py'

dist:
	@echo "Building distribution..."
	. venv/bin/activate; python setup.py sdist

clean:
	@echo "Cleaning up..."
	rm -rf dist

release:
	@echo "Releasing to PyPI..."
	twine upload dist/*

