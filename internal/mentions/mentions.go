package mentions

import (
	"context"
	"regexp"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/attachments"
	"github.com/needmore/bc4/internal/utils"
)

// mentionRe matches @Name and @First.Last mentions in rich text content.
// Mentions must appear at the start of the string, after > (blockquotes),
// or after whitespace.
var mentionRe = regexp.MustCompile(`(?:^|[>\s])(@[\w]+(?:\.[\w]+)*)`)

// Resolve finds @mentions in rich text content and replaces them with
// Basecamp bc-attachment tags. It uses the project's people list to
// resolve mention identifiers to their AttachableSGID values.
//
// Supports @FirstName and @First.Last syntax. Returns the content
// unchanged if no mentions are found.
func Resolve(ctx context.Context, richContent string, client api.APIClient, projectID string) (string, error) {
	submatches := mentionRe.FindAllStringSubmatch(richContent, -1)
	if len(submatches) == 0 {
		return richContent, nil
	}

	resolver := utils.NewUserResolver(client, projectID)

	// Extract capture group (the @mention) and convert @First.Last to "First Last"
	mentions := make([]string, len(submatches))
	identifiers := make([]string, len(submatches))
	for i, sm := range submatches {
		mentions[i] = sm[1]
		identifiers[i] = strings.ReplaceAll(strings.TrimPrefix(sm[1], "@"), ".", " ")
	}

	people, err := resolver.ResolvePeople(ctx, identifiers)
	if err != nil {
		return "", err
	}

	for i, match := range mentions {
		tag := attachments.BuildTag(people[i].AttachableSGID)
		richContent = strings.Replace(richContent, match, tag, 1)
	}

	return richContent, nil
}
