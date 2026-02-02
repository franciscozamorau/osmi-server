#!/bin/bash
echo "Corrigiendo imports de repositorios PostgreSQL..."

# Lista de archivos a corregir
find internal/infrastructure/repositories/postgres -name "*.go" | while read file; do
    echo "Corrigiendo: $file"
    
    # Reemplazar imports incorrectos
    sed -i 's|github.com/franciscozamorau/osmi-server/repositories/postgres/helpers/|github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres/helpers/|g' "$file"
    
    # También corregir otras referencias si existen
    sed -i 's|repositories/postgres/helpers/|internal/infrastructure/repositories/postgres/helpers/|g' "$file"
    
    echo "  ✓ $file corregido"
done

echo ""
echo "Verificando correcciones..."
grep -r "repositories/postgres/helpers" internal/infrastructure/repositories/postgres/ || echo "✅ Todos los imports corregidos"
