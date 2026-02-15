#!/bin/bash
# Generate API documentation and frontend client code
# This script runs swag to generate OpenAPI docs, then generates the frontend client

echo ""
echo "=== Generating Backend OpenAPI Documentation ==="
echo ""

cd backend
echo "Running swag init in: $(pwd)"

go run github.com/swaggo/swag/cmd/swag init -g main.go

if [ $? -ne 0 ]; then
    echo "Error: Swag generation failed!"
    exit 1
fi

echo ""
echo "✓ Swagger docs generated successfully"
echo ""

cd ..

echo "=== Generating Frontend Client Code ==="
echo ""

cd frontend
echo "Running npm generate in: $(pwd)"

npm run generate

if [ $? -ne 0 ]; then
    echo "Error: Frontend client generation failed!"
    exit 1
fi

echo ""
echo "✓ Frontend client generated successfully"
echo ""

cd ..

echo ""
echo "=== All documentation generated successfully! ==="
echo ""
