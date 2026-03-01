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
	indexMatches := mentionRe.FindAllStringSubmatchIndex(richContent, -1)
	if len(indexMatches) == 0 {
		return richContent, nil
	}

	resolver := utils.NewUserResolver(client, projectID)

	// Extract capture group (the @mention) and convert @First.Last to "First Last"
	identifiers := make([]string, len(indexMatches))
	for i, loc := range indexMatches {
		// loc[2]:loc[3] is the capture group (the @mention)
		mention := richContent[loc[2]:loc[3]]
		identifiers[i] = strings.ReplaceAll(strings.TrimPrefix(mention, "@"), ".", " ")
	}

	people, err := resolver.ResolvePeople(ctx, identifiers)
	if err != nil {
		return "", err
	}

	// Replace from end to start so earlier indices remain valid
	for i := len(indexMatches) - 1; i >= 0; i-- {
		tag := attachments.BuildTag(people[i].AttachableSGID)
		richContent = richContent[:indexMatches[i][2]] + tag + richContent[indexMatches[i][3]:]
	}

	return richContent, nil
}
