package api

import (
	"github.com/ProNexus-Startup/ProNexus/backend/database"
)

// initializeHandlers creates and returns all handlers organized in a routeHandlers struct
func initializeHandlers(database database.Database, backendPassword string) *routeHandlers {
	return &routeHandlers{
		projectHandler:  newProjectHandler(database.ProjectRepo(), database.ProjectTagRepo()),
		blogPostHandler: newBlogPostHandler(database.BlogPostRepo(), database.BlogTagRepo()),
	}
}
