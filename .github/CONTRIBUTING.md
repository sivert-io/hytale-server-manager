# Contributing to Hytale Server Manager

Thank you for your interest in contributing to HSM! 🎉

## 🚀 Quick Start

1. **Fork & Clone**

   ```bash
   git clone https://github.com/YOUR_USERNAME/hytale-server-manager.git
   cd hytale-server-manager
   ```

2. **Install Dependencies**

   ```bash
   # Ensure Go 1.19+ is installed
   go version
   
   # Download dependencies
   go mod download
   ```

3. **Build the Binary**

   ```bash
   # Build locally
   go build -ldflags="-s -w" -o ./hsm ./src/cmd/hytale-tui
   
   # Or install globally (requires sudo)
   sudo ./install.sh
   ```

4. **Run the TUI**

   ```bash
   # If built locally
   sudo ./hsm
   
   # If installed globally
   sudo hsm
   ```

## 📝 Development Guidelines

- ✅ Write clear commit messages
- ✅ Test your changes locally
- ✅ Update documentation if needed
- ✅ Follow Go code style conventions
- ✅ Keep PRs focused on one feature/fix
- ✅ Ensure the TUI builds without errors

## 🧪 Testing

Before submitting a PR, please:

1. Build the binary successfully:
   ```bash
   go build ./src/cmd/hytale-tui
   ```

2. Run the TUI and verify basic functionality:
   ```bash
   sudo ./hsm
   ```

3. Test any new features you've added

## 📦 Project Structure

```
hytale-server-manager/
├── src/
│   ├── cmd/hytale-tui/    # TUI entry point
│   └── internal/
│       ├── tui/           # TUI layer (user interface)
│       └── hytale/        # Backend layer (server management)
├── scripts/               # Server scripts
├── tools/                 # Helper scripts
└── docs/                  # Documentation
```

## 🐛 Reporting Issues

Found a bug? Please [open an issue](https://github.com/sivert-io/hytale-server-manager/issues/new) with:

- Clear description
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version, Java version, etc.)

## 🙏 Community Requests

Need help testing something or getting feedback? Use the **Community Request** issue template! This is perfect for:

- Features that require multiple servers to test
- Cross-platform compatibility testing
- Getting user experience feedback
- Performance testing with real-world scenarios

**Contributors who help with Community Requests will be recognized and credited!** 🏆

## 💬 Questions?

- [GitHub Discussions](https://github.com/sivert-io/hytale-server-manager/discussions) - Ask questions
- [GitHub Issues](https://github.com/sivert-io/hytale-server-manager/issues) - Report bugs or request features

## 📖 Code of Conduct

Be respectful and constructive. We're all here to build something awesome for the Hytale community! 🎮
