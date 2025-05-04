# Contributing to KubeGPT

Thank you for your interest in contributing to KubeGPT! This document provides guidelines and instructions for contributing to this project.

## Development Setup

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/yourusername/kubegpt.git
   cd kubegpt
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Build the project:
   ```bash
   make build
   ```

## Development Workflow

1. Create a new branch for your feature or bugfix:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and write tests if applicable

3. Run tests:
   ```bash
   make test
   ```

4. Format your code:
   ```bash
   go fmt ./...
   ```

5. Commit your changes with a descriptive commit message:
   ```bash
   git commit -m "Add feature: your feature description"
   ```

6. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

7. Create a Pull Request from your fork to the main repository

## Code Style

- Follow standard Go code style and conventions
- Use meaningful variable and function names
- Add comments for complex logic
- Write unit tests for new functionality

## Pull Request Process

1. Ensure your code passes all tests
2. Update documentation if necessary
3. Add your changes to the CHANGELOG.md file
4. Your PR will be reviewed by maintainers who may request changes
5. Once approved, your PR will be merged

## Adding New Commands

When adding a new command to KubeGPT:

1. Create a new file in the `cmd` directory
2. Implement the command using the Cobra framework
3. Add the command to the root command in `cmd/root.go`
4. Add appropriate tests
5. Update documentation

## Testing with Mock Mode

For development and testing without Amazon Q, you can use the mock mode:

```bash
export KUBEGPT_MOCK_AI=true
make run
```

## Reporting Issues

When reporting issues, please include:

- KubeGPT version
- Kubernetes version
- Operating system
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Any relevant logs or error messages

## Feature Requests

Feature requests are welcome! Please provide:

- A clear description of the feature
- The use case for the feature
- Any ideas for implementation

Thank you for contributing to KubeGPT!