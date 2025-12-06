package api

import (
	"github.com/go-chi/chi/v5"
)

// setupFrontendRoutes sets up all routes with authentication
func setupFrontendRoutes(r chi.Router, handlers *routeHandlers, authMiddleware authMiddleware) {
	// Authenticated routes
	r.Group(func(r chi.Router) {
		//r.Use(authMiddleware.authenticate)
		r.Use(ColoredHTTPLoggingMiddleware)

		// Project Handler endpoints
		r.Get("/projects", handlers.projectHandler.getAllProjects())
		r.Get("/project/{projectID}", handlers.projectHandler.getProject())
		r.Post("/project", handlers.projectHandler.createProject())
		r.Put("/project/{projectID}", handlers.projectHandler.updateProject())
		r.Delete("/project/{projectID}", handlers.projectHandler.deleteProject())

		// Blog Post Handler endpoints
		r.Get("/blog-posts", handlers.blogPostHandler.getAllBlogPosts())
		r.Get("/blog-post/{blogPostID}", handlers.blogPostHandler.getBlogPost())
		r.Post("/blog-post", handlers.blogPostHandler.createBlogPost())
		r.Put("/blog-post/{blogPostID}", handlers.blogPostHandler.updateBlogPost())
		r.Delete("/blog-post/{blogPostID}", handlers.blogPostHandler.deleteBlogPost())
	})
}
