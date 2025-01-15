# Business Canvas

A desktop application built with Go and Fyne framework for creating and managing business model canvases. This tool helps entrepreneurs and business planners visualize and develop their business models using the Business Model Canvas framework.

## Project Structure
```
.
├── bundled.go
├── FyneApp.toml
├── go.mod
├── go.sum
├── icon.png
├── README.md
└── main.go
```

## Prerequisites

- Go 1.16 or later
- Fyne dependencies

### Installing Fyne Dependencies

#### Ubuntu/Debian
```bash
sudo apt-get install gcc libgl1-mesa-dev xorg-dev
```

#### Fedora
```bash
sudo dnf install gcc libXcursor-devel libXrandr-devel mesa-libGL-devel libXi-devel libXinerama-devel libXxf86vm-devel
```

#### macOS
```bash
brew install go gcc
```

#### Windows
- Install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/)
- Install [Go](https://golang.org/dl/)

## Installation

1. Clone the repository
```bash
git clone <repository-url>
cd business-canvas
```

2. Install Go dependencies
```bash
go mod download
```

3. Run the application
```bash
go run main.go
```

## Features

- Interactive Business Model Canvas with 9 key sections
- Auto-save functionality
- Dark/Light theme options
- Export to PDF
- Version history
- Progress tracking
- Real-time validation

## Usage

### Keyboard Shortcuts
- `Ctrl + S`: Save canvas
- `Ctrl + O`: Open canvas
- `Ctrl + Z`: Undo
- `Ctrl + Y`: Redo
- `Ctrl + P`: Export to PDF
- `Ctrl + C`: Copy
- `Ctrl + V`: Paste
- `Ctrl + X`: Cut

### Canvas Sections
- Key Partners
- Key Activities
- Key Resources
- Value Propositions
- Customer Relationships
- Channels
- Customer Segments
- Cost Structure
- Revenue Streams

## Building

To build the application:

```bash
go build
```

This will create an executable in your project directory.

## License

[MIT License](LICENSE)
