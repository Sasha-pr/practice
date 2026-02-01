
#!/bin/bash

echo "Installing swag command..."
go install github.com/swaggo/swag/cmd/swag@latest

echo "Generating Swagger documentation..."
swag init -g main.go -o ./docs --parseDependency --parseInternal

echo "Documentation generated in ./docs directory"
echo "To view documentation, start the server and visit:"
echo "http://localhost:8080/swagger/index.html"
