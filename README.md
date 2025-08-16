# PRism

**PRism** is a modern, interactive command-line utility for viewing and managing your GitHub Pull Requests that require your review. It provides a fast, TUI-based experience with filtering, scrolling, and detailed PR views—all from your terminal.

---

## Features

- 🚦 **See all PRs requesting your review**  
  Instantly fetches open PRs where you are requested as a reviewer (not via team).
- 🖱️ **Interactive TUI**  
  Navigate with arrow keys, filter with `/`, and select PRs with `Enter`.
- 🔍 **Detailed PR View**  
  Press `Enter` to see the full PR description, scroll with `j/k` or arrow keys.
- ↩️ **Quick Return**  
  Press `q` or `esc` to return to the list view.
- 📏 **Auto-resizing**  
  The UI adapts to your terminal size for optimal viewing.
- ⚠️ **Error Reporting**  
  Notifies you if any PRs could not be fetched due to API errors.

---

## Installation

1. **Clone the repository:**
   ```sh
   git clone https://github.com/yourusername/prism.git
   cd prism
   ```

2. **Build the binary:**
   ```sh
   go build -o prism
   ```

3. **Set your GitHub token:**
   ```sh
   export GITHUB_TOKEN=your_github_token
   ```

---

## Usage

Simply run:

```sh
./prism
```

- Use <kbd>↑</kbd>/<kbd>↓</kbd> or <kbd>j</kbd>/<kbd>k</kbd> to navigate.
- Press <kbd>/</kbd> to filter PRs.
- Press <kbd>Enter</kbd> to view PR details.
- In detail view, scroll with <kbd>j</kbd>/<kbd>k</kbd> or <kbd>↓</kbd>/<kbd>↑</kbd>.
- Press <kbd>q</kbd> or <kbd>esc</kbd> to return to the list.

---

## Requirements

- Go 1.18+
- A valid GitHub personal access token with `repo` scope (`GITHUB_TOKEN` environment variable)

---

## Why PRism?

- **Fast:** No more clicking through GitHub notifications.
- **Focused:** Only see PRs that need _your_ review.
- **Beautiful:** Uses [Bubble Tea](https://github.com/charmbracelet/bubbletea) for a modern TUI experience.

---

## Contributing

Contributions are welcome! Please open issues or PRs for bugs, features, or suggestions.

---

## License

MIT License

---

**PRism** — _A new perspective on your Pull Requests!_
