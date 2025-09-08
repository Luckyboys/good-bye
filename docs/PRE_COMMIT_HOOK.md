# Git Pre-commit Hook Setup

This project includes a comprehensive pre-commit hook that runs various code quality checks before allowing commits.

## Features

The pre-commit hook includes the following checks:

### Basic Go Checks
- **go fmt**: Code formatting
- **go vet**: Bug detection and suspicious constructs
- **golint**: Style guide compliance

### Modernize Checks
- **gosimple**: Code simplification suggestions
- **staticcheck**: Advanced static analysis
- **unused**: Unused constants, variables, functions, and types
- **gosec**: Security vulnerability detection
- **errcheck**: Unchecked error detection
- **ineffassign**: Inefficient assignment detection
- **prealloc**: Pre-allocation optimization suggestions
- **unparam**: Unused parameter detection
- **nakedret**: Naked return function detection
- **exhaustive**: Enum switch exhaustiveness checking
- **exptostd**: Experimental to standard library conversion suggestions
- **copyloopvar**: Loop variable copying detection
- **errname**: Error naming convention checking
- **errorlint**: Modern error handling practices
- **durationcheck**: Duration multiplication checking
- **dogsled**: Too many blank identifiers in assignments
- **dupl**: Code duplication detection
- **misspell**: Common misspellings detection

### Dependency Management
- **go mod tidy**: Automatic dependency management when go.mod/go.sum changes

### Documentation Checks
- **Markdown formatting**: Ensures documentation files are properly formatted
- **File existence**: Verifies required documentation files exist

## Installation

### Quick Setup
Run the setup script to automatically install the pre-commit hook:

```bash
./scripts/setup-pre-commit.sh
```

### Manual Setup
1. Ensure golangci-lint is installed:
   ```bash
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

2. Copy the pre-commit hook:
   ```bash
   cp scripts/pre-commit-hook .git/hooks/pre-commit
   chmod +x .git/hooks/pre-commit
   ```

## Usage

The pre-commit hook runs automatically when you attempt to commit changes. It will:

1. Check if any Go files are staged for commit
2. Run all enabled checks on the staged files
3. If any check fails, the commit will be aborted
4. Provide detailed error messages to help fix issues

### Example Output

```
🔍 Running pre-commit checks...
🔍 Found Go files to check:
   - src/email/retry_manager.go

🔍 Running go fmt...
✅ go fmt check passed
🔍 Running go vet...
✅ go vet check passed
🔍 Running golint...
✅ golint check passed
🔍 Running modernize checks...
src/email/retry_manager.go:201:53: interface{} can be replaced by any (modernize)
❌ modernize checks found issues
🔍 Please fix the issues reported by modernize checks
```

## Bypassing the Hook

If you need to bypass the pre-commit hook (not recommended), you can use:

```bash
git commit --no-verify
```

However, this should only be used in exceptional circumstances.

## Troubleshooting

### golangci-lint Not Found
If you get an error about golangci-lint not being found:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Hook Not Executable
If the hook doesn't run, make sure it's executable:

```bash
chmod +x .git/hooks/pre-commit
```

### Specific Linter Issues
If a particular linter is causing problems, you can temporarily disable it by modifying the `--enable` flags in the pre-commit hook.

## Configuration

The pre-commit hook is configured to run a specific set of linters focused on code modernization and quality. You can modify the linter selection by editing the `.git/hooks/pre-commit` file.

### Adding New Linters
To add new linters, add them to the `--enable` list:

```bash
golangci-lint run --disable-all --enable=gosimple --enable=staticcheck --enable=your-new-linter ./...
```

### Disabling Linters
To disable a linter, remove it from the `--enable` list.

## Benefits

Using the pre-commit hook provides several benefits:

1. **Consistent Code Quality**: Ensures all code meets the same standards
2. **Early Issue Detection**: Catches issues before they reach the repository
3. **Modern Code Practices**: Encourages modern Go idioms and practices
4. **Security**: Detects potential security vulnerabilities
5. **Performance**: Identifies performance optimization opportunities
6. **Reduced Review Time**: Reduces time spent on code style issues in review

## Maintenance

The pre-commit hook setup script and documentation should be kept up to date with any changes to the project's tooling or requirements.

## CI/CD Integration

The pre-commit hook is integrated with the project's CI/CD pipeline:

### GitHub Actions
- The same checks run in the CI/CD pipeline
- Ensures consistency between local development and CI
- Fails builds if any check fails

### Makefile Integration
```bash
# Run all checks locally
make check

# Run specific checks
make fmt    # go fmt
make lint   # golangci-lint
make vet    # go vet
```

## Performance Considerations

### Optimization Tips
1. **Incremental Checks**: Only checks staged files, not the entire project
2. **Parallel Execution**: Some checks can run in parallel
3. **Caching**: Dependency caching for faster execution

### Large Project Considerations
For large projects, consider:
- Running checks only on modified files
- Using caching mechanisms
- Staggering check execution

## Troubleshooting Advanced Issues

### Memory Issues
If golangci-lint runs out of memory on large files:
```bash
# Increase memory limit
golangci-lint run --memory-limit=2048 ./...
```

### Timeout Issues
If checks timeout on large codebases:
```bash
# Increase timeout
golangci-lint run --timeout=5m ./...
```

### False Positives
Some linters may have false positives:
- Use `//nolint` comments to disable specific linters for specific lines
- Report false positives to the linter's issue tracker
- Consider disabling problematic linters if necessary

## Best Practices

### Code Quality
- **Don't bypass**: Avoid using `--no-verify` unless absolutely necessary
- **Fix issues immediately**: Address all issues before committing
- **Review warnings**: Even if warnings don't fail the build, review them

### Team Collaboration
- **Consistent environment**: Ensure all team members use the same tool versions
- **Documentation**: Keep documentation up to date with hook changes
- **Training**: Help new team members understand the hook requirements

### Continuous Improvement
- **Regular updates**: Keep linters and tools updated
- **Feedback loop**: Report issues and suggest improvements
- **Metrics**: Track hook effectiveness and failure rates