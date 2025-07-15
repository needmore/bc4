# Modular API Client Usage Guide

This guide shows how to use the new modular API clients that have been created alongside the existing monolithic client.

## Current Status

The modular clients have been implemented but are not yet integrated into the command implementations. They exist alongside the original monolithic `api.Client` to maintain backward compatibility.

## Available Modular Clients

- `project.Client` - Project-related operations
- `todo.Client` - Todo list and todo item operations  
- `campfire.Client` - Campfire chat operations
- `card.Client` - Card table (kanban board) operations
- `people.Client` - People/user operations

## Usage Example

To use the modular clients, you would create instances of each client you need:

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/needmore/bc4/internal/api"
    "github.com/needmore/bc4/internal/api/project"
    "github.com/needmore/bc4/internal/api/todo"
)

func main() {
    // Create base client
    base := api.NewBaseClient("accountID", "accessToken")
    
    // Create modular clients
    projectClient := project.NewClient(base)
    todoClient := todo.NewClient(base, projectClient)
    
    // Use the clients
    ctx := context.Background()
    projects, err := projectClient.GetProjects(ctx)
    if err != nil {
        panic(err)
    }
    
    for _, p := range projects {
        fmt.Printf("Project: %s\n", p.Name)
    }
}
```

## Benefits of Modular Design

1. **Better testability** - Each client can be mocked independently
2. **Clearer dependencies** - It's obvious which operations depend on others
3. **Smaller interfaces** - Each interface only contains related methods
4. **Easier to extend** - New resource types can be added without modifying existing code

## Migration Path

The existing `api.Client` continues to work as before. To migrate:

1. Replace `api.NewClient()` with individual client creation
2. Update method calls to use the appropriate client
3. Update mocks to implement the smaller interfaces

## Next Steps

1. Update command implementations to use modular clients
2. Update mock implementations to support new interfaces
3. Remove methods from monolithic client once migration is complete