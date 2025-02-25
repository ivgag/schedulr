
#!/bin/bash
# Run this script from the workspace root directory

# Find all directories containing a go.mod file
modules=$(find .. -type f -name 'go.mod' -exec dirname {} \;)

# Iterate over each module directory and run go mod tidy
for mod in $modules; do
    echo "Running go mod tidy in $mod"
    (cd "$mod" && go mod tidy)
done