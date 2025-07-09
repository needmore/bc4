# Table Formatting in bc4

This document describes the table formatting capabilities in bc4, inspired by the GitHub CLI (gh).

## Features

### 1. Multiple Output Formats

All list commands now support multiple output formats via the `--format` flag:

- `table` (default): Human-readable table with borders and color
- `json`: JSON output for scripting
- `tsv`: Tab-separated values for easy parsing

Example:
```bash
bc4 project list --format=json
bc4 account list --format=tsv
bc4 todo list --format=table
```

### 2. Responsive Column Widths

Tables automatically adjust column widths based on terminal size:

- **Minimum widths**: Each column has a minimum width to ensure readability
- **Maximum widths**: Columns can have maximum widths to prevent excessive stretching
- **Flexible columns**: Certain columns (like descriptions) can expand to use available space
- **Proportional shrinking**: When terminal is too narrow, columns shrink proportionally

### 3. Terminal Detection

The system automatically detects whether output is going to a terminal:

- **TTY mode**: Full formatting with colors, borders, and truncation
- **Non-TTY mode**: Tab-separated output without formatting (for piping)

### 4. Environment Variables

Respect standard terminal environment variables:

- `NO_COLOR`: Disables all color output
- `CLICOLOR=0`: Also disables color output
- `BC4_FORCE_TTY` or `FORCE_COLOR`: Forces color output even in non-TTY

## Implementation Details

### Column Width Algorithm

The `CalculateColumnWidths` function in `internal/ui/table.go` implements a smart algorithm:

1. Start with preferred widths for each column
2. Apply minimum width constraints
3. Apply maximum width constraints
4. If content fits in terminal, distribute extra space to flexible columns
5. If content is too wide, proportionally shrink all columns

### Output Formatting

The `internal/ui/output.go` module provides:

- `OutputConfig`: Configuration for output formatting
- `TableWriter`: Interface for writing tabular data
- Format conversion between table, JSON, and TSV

### Integration with Lipgloss

Tables use the Charmbracelet lipgloss library for styling:

- Border styles (normal, markdown, ASCII)
- Color support with automatic degradation
- Header styling with bold and colors
- Responsive width handling

## Usage Examples

### Basic Table Output
```bash
bc4 project list
```

### JSON Output for Scripting
```bash
bc4 account list --format=json | jq '.[].Name'
```

### TSV Output for Spreadsheets
```bash
bc4 todo list --format=tsv > todos.tsv
```

### Disable Colors
```bash
NO_COLOR=1 bc4 project list
```

## Future Enhancements

1. **Column sorting**: Add ability to sort by different columns
2. **Filtering**: Add ability to filter rows based on criteria
3. **Custom column selection**: Allow users to choose which columns to display
4. **Pager integration**: Automatically page long tables
5. **Export formats**: Add CSV, Markdown table formats