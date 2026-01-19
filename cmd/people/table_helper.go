package people

import (
	"os"
	"strconv"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/ui/tableprinter"
)

// renderPeopleTable renders a list of people in a formatted table
func renderPeopleTable(people []api.Person, includeRole bool) error {
	// Create new GitHub CLI-style table
	table := tableprinter.New(os.Stdout)

	// Add headers dynamically based on TTY mode
	if table.IsTTY() {
		if includeRole {
			table.AddHeader("ID", "NAME", "EMAIL", "TITLE", "ROLE")
		} else {
			table.AddHeader("ID", "NAME", "EMAIL", "TITLE")
		}
	} else {
		if includeRole {
			table.AddHeader("ID", "NAME", "EMAIL", "TITLE", "ROLE", "COMPANY")
		} else {
			table.AddHeader("ID", "NAME", "EMAIL", "TITLE", "COMPANY")
		}
	}

	// Add people to table
	for _, person := range people {
		// Determine role
		role := "member"
		if person.Owner {
			role = "owner"
		} else if person.Admin {
			role = "admin"
		}

		// Add ID field
		table.AddIDField(strconv.FormatInt(person.ID, 10), role)

		// Add name
		cs := table.GetColorScheme()
		table.AddField(person.Name, cs.Bold)

		// Add email
		table.AddField(person.EmailAddress)

		// Add title
		table.AddField(person.Title, cs.Muted)

		// Add role with color (if includeRole is true)
		if includeRole {
			switch role {
			case "owner":
				table.AddField(role, cs.Magenta)
			case "admin":
				table.AddField(role, cs.Cyan)
			default:
				table.AddField(role, cs.Muted)
			}
		}

		// Add company only for non-TTY
		if !table.IsTTY() {
			companyName := ""
			if person.Company != nil {
				companyName = person.Company.Name
			}
			table.AddField(companyName)
		}

		table.EndRow()
	}

	return table.Render()
}
