#!/bin/bash

# Setup script for git pre-commit hooks
# This script installs the modernize-enabled pre-commit hook

set -e

echo "🔧 Setting up git pre-commit hooks..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}🔧 $1${NC}"
}

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    print_error "Not in a git repository"
    exit 1
fi

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    print_info "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    print_success "golangci-lint installed successfully"
fi

# Create the pre-commit hook
print_info "Creating pre-commit hook..."

cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash

# Git pre-commit hook for Go projects with modernize checks
# This hook runs go fmt, go vet, golint, and modernize checks before allowing commit

set -e

echo "🔍 Running pre-commit checks..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${YELLOW}🔍 $1${NC}"
}

# Check if this is a Go project
if [ ! -f "go.mod" ]; then
    print_warning "Not a Go module project, skipping Go checks"
    exit 0
fi

# Get list of staged Go files
STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [ -z "$STAGED_GO_FILES" ]; then
    print_success "No Go files staged, skipping checks"
    exit 0
fi

print_info "Found Go files to check:"
for file in $STAGED_GO_FILES; do
    echo "   - $file"
done

echo ""

# Run go fmt
print_info "Running go fmt..."
FMT_OUTPUT=$(gofmt -l $STAGED_GO_FILES 2>&1 || true)
if [ -n "$FMT_OUTPUT" ]; then
    print_error "go fmt found issues in the following files:"
    echo "$FMT_OUTPUT" | sed 's/^/   /'
    echo ""
    print_info "Please run 'go fmt ./...' to fix formatting issues"
    exit 1
fi
print_success "go fmt check passed"

# Run go vet
print_info "Running go vet..."
if ! go vet ./... 2>/dev/null; then
    print_error "go vet found issues"
    print_info "Please fix the issues reported by go vet"
    exit 1
fi
print_success "go vet check passed"

# Run golint
print_info "Running golint..."
LINT_ERRORS=0
for file in $STAGED_GO_FILES; do
    if [ -f "$file" ]; then
        LINT_OUTPUT=$(golint "$file" 2>&1 || true)
        if [ -n "$LINT_OUTPUT" ]; then
            print_error "golint found issues in $file:"
            echo "$LINT_OUTPUT" | sed 's/^/   /'
            echo ""
            LINT_ERRORS=1
        fi
    fi
done

if [ $LINT_ERRORS -ne 0 ]; then
    print_error "golint found issues that need to be fixed"
    print_info "Please fix the issues reported by golint"
    exit 1
fi
print_success "golint check passed"

# Run modernize checks
print_info "Running modernize checks..."
if ! golangci-lint run --disable-all --enable=gosimple --enable=staticcheck --enable=unused --enable=gosec --enable=errcheck --enable=ineffassign --enable=prealloc --enable=unparam --enable=nakedret --enable=exhaustive --enable=exptostd --enable=copyloopvar --enable=errname --enable=errorlint --enable=durationcheck --enable=dogsled --enable=dupl --enable=misspell ./... 2>/dev/null; then
    print_error "modernize checks found issues"
    print_info "Please fix the issues reported by modernize checks"
    exit 1
fi
print_success "modernize checks passed"

# Check for go mod tidy if go.mod is modified
if git diff --cached --name-only | grep -q "go.mod\|go.sum"; then
    print_info "go.mod or go.sum modified, running go mod tidy..."
    if ! go mod tidy 2>/dev/null; then
        print_error "go mod tidy failed"
        exit 1
    fi
    print_success "go mod tidy completed"
    
    # Check if go.mod or go.sum has changes after tidy
    if [ -n "$(git diff --name-only go.mod go.sum 2>/dev/null || true)" ]; then
        print_warning "go mod tidy made changes to go.mod or go.sum"
        print_info "Please add these changes to your commit:"
        echo "   git add go.mod go.sum"
        echo ""
        print_info "Then run the commit again"
        exit 1
    fi
fi

echo ""
print_success "🎉 All pre-commit checks passed!"
echo "   ✅ go fmt"
echo "   ✅ go vet"
echo "   ✅ golint"
echo "   ✅ modernize"
echo ""

exit 0
EOF

# Make the hook executable
chmod +x .git/hooks/pre-commit

print_success "Git pre-commit hook setup completed!"
echo ""
print_info "The pre-commit hook now includes:"
echo "   ✅ go fmt - Code formatting"
echo "   ✅ go vet - Bug detection"
echo "   ✅ golint - Style guide compliance"
echo "   ✅ modernize - Code modernization checks"
echo ""
print_info "Modernize checks include:"
echo "   • gosimple - Code simplification"
echo "   • staticcheck - Static analysis"
echo "   • unused - Unused code detection"
echo "   • gosec - Security checks"
echo "   • errcheck - Error handling"
echo "   • ineffassign - Inefficient assignments"
echo "   • prealloc - Pre-allocation optimization"
echo "   • unparam - Unused parameters"
echo "   • nakedret - Naked return detection"
echo "   • exhaustive - Enum exhaustiveness"
echo "   • exptostd - Experimental to std conversion"
echo "   • copyloopvar - Loop variable copying"
echo "   • errorlint - Error handling modernization"
echo "   • And more..."
echo ""
print_success "Pre-commit hook is now active and will run on all commits!"