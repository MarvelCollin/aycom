@echo off
echo Manually generating Swagger documentation...

if exist "%~dp0\docs\swagger.yaml" (
    echo Found swagger.yaml file

    rem Try to convert YAML to JSON using Python
    python -c "import sys, yaml, json; json.dump(yaml.safe_load(open('./docs/swagger.yaml', 'r')), open('./docs/swagger.json', 'w'), indent=2)" 2>nul
    if %ERRORLEVEL% EQU 0 (
        echo Successfully converted YAML to JSON using Python
        goto Success
    )

    rem If Python failed, try PowerShell
    powershell -Command "ConvertFrom-Yaml (Get-Content -Raw ./docs/swagger.yaml) | ConvertTo-Json -Depth 100 | Out-File -Encoding utf8 ./docs/swagger.json" 2>nul
    if %ERRORLEVEL% EQU 0 (
        echo Successfully converted YAML to JSON using PowerShell
        goto Success
    )

    echo Failed to convert YAML to JSON
    echo Please install Python with PyYAML package or PowerShell with yaml module
    exit /b 1
) else (
    echo Error: swagger.yaml file not found
    echo Make sure the file exists at %~dp0\docs\swagger.yaml
    exit /b 1
)

:Success
echo Swagger documentation generated successfully at ./docs/swagger.json
echo Access the Swagger UI at: http://localhost:8083/swagger/index.html
