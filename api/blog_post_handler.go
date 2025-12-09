package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rpupo63/unified-personal-site-backend/database"
	"github.com/rpupo63/unified-personal-site-backend/errs"
	"github.com/rpupo63/unified-personal-site-backend/models"
	"github.com/rpupo63/unified-personal-site-backend/services"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type blogPostHandler struct {
	responder    Responder
	logger       zerolog.Logger
	blogPostRepo *database.BlogPostRepo
	blogTagRepo  *database.BlogTagRepo
}

func newBlogPostHandler(blogPostRepo *database.BlogPostRepo, blogTagRepo *database.BlogTagRepo) blogPostHandler {
	logger := log.With().Str("handlerName", "blogPostHandler").Logger()

	return blogPostHandler{
		responder:    NewResponder(logger),
		logger:       logger,
		blogPostRepo: blogPostRepo,
		blogTagRepo:  blogTagRepo,
	}
}

// BlogPostWithTags represents a blog post with its tags
type BlogPostWithTags struct {
	BlogPost models.BlogPost  `json:"blogPost"`
	Tags     []models.BlogTag `json:"tags"`
}

// BlogPostCollectionWithTags represents multiple blog posts with their tags
type BlogPostCollectionWithTags struct {
	BlogPosts []BlogPostWithTags `json:"blogPosts"`
	Total     int                `json:"total,omitempty"`
}

// getAllBlogPosts retrieves all blog posts with their tags
// @Summary Get all blog posts
// @Description Retrieves all blog posts from the database with their associated tags
// @Tags Blog Posts
// @Accept json
// @Produce json
// @Success 200 {object} BlogPostCollectionWithTags "List of blog posts with tags"
// @Failure 500 {object} errs.ErrorResponse "Internal Server Error - Error fetching blog posts"
// @Router /blog-posts [get]
func (h blogPostHandler) getAllBlogPosts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authentication handled by middleware

		blogPosts, err := h.blogPostRepo.FindAll()
		if err != nil {
			h.responder.WriteError(w, wrapDatabaseError("find blog posts", "blog_posts", err))
			return
		}

		// Convert to BlogPostWithTags format
		var blogPostsWithTags []BlogPostWithTags
		for _, blogPost := range blogPosts {
			blogPostsWithTags = append(blogPostsWithTags, BlogPostWithTags{
				BlogPost: *blogPost,
				Tags:     blogPost.Tags,
			})
		}

		response := BlogPostCollectionWithTags{
			BlogPosts: blogPostsWithTags,
			Total:     len(blogPostsWithTags),
		}

		h.responder.WriteJSON(w, response)
	}
}

// getBlogPost retrieves a specific blog post by ID with its tags
// @Summary Get blog post
// @Description Retrieves detailed information about a specific blog post by ID with its tags
// @Tags Blog Posts
// @Accept json
// @Produce json
// @Param blogPostID path string true "Blog Post ID" format(uuid)
// @Success 200 {object} BlogPostWithTags "Blog post details with tags"
// @Failure 400 {object} errs.ErrorResponse "Bad Request - Invalid blogPostID"
// @Failure 404 {object} errs.ErrorResponse "Not Found - Blog post not found"
// @Failure 500 {object} errs.ErrorResponse "Internal Server Error - Error fetching blog post"
// @Router /blog-post/{blogPostID} [get]
func (h blogPostHandler) getBlogPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authentication handled by middleware

		blogPostIDStr := chi.URLParam(r, "blogPostID")
		if blogPostIDStr == "" {
			h.responder.WriteError(w, errs.NewBadRequestError("missing blogPostID"))
			return
		}

		blogPostID, err := uuid.Parse(blogPostIDStr)
		if err != nil {
			h.responder.WriteError(w, errs.NewBadRequestError("invalid blogPostID"))
			return
		}

		blogPost, err := h.blogPostRepo.FindByID(blogPostID)
		if err != nil {
			h.responder.WriteError(w, wrapDatabaseError("find blog post", "blog_post", err))
			return
		}

		if blogPost == nil {
			h.responder.WriteError(w, errs.NewNotFoundError("blog post not found"))
			return
		}

		response := BlogPostWithTags{
			BlogPost: *blogPost,
			Tags:     blogPost.Tags,
		}

		h.responder.WriteJSON(w, response)
	}
}

// createBlogPost creates a new blog post
// @Summary Create blog post
// @Description Creates a new blog post in the database and posts it to all configured social media platforms
// @Tags Blog Posts
// @Accept json
// @Produce json
// @Param blogPost body models.BlogPost true "Blog post data"
// @Param mainImageURL query string false "Main image URL for Substack posting"
// @Success 201 {object} BlogPostWithTags "Created blog post with tags"
// @Failure 400 {object} errs.ErrorResponse "Bad Request - Invalid blog post data"
// @Failure 500 {object} errs.ErrorResponse "Internal Server Error - Error creating blog post"
// @Router /blog-post [post]
func (h blogPostHandler) createBlogPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authentication handled by middleware

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to read request body")
			h.responder.WriteError(w, errs.NewBadRequestError("failed to read request body"))
			return
		}

		var blogPost models.BlogPost
		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&blogPost); err != nil {
			h.logger.Error().Err(err).Str("body", string(bodyBytes)).Msg("Failed to decode blog post request body")
			h.responder.WriteError(w, errs.NewBadRequestError("malformed request body"))
			return
		}

		if blogPost.Title == "" {
			h.responder.WriteError(w, errs.NewBadRequestError("title is required"))
			return
		}

		if blogPost.Content == "" {
			h.responder.WriteError(w, errs.NewBadRequestError("content is required"))
			return
		}

		// Set DateAdded if not provided
		if blogPost.DateAdded.IsZero() {
			blogPost.DateAdded = time.Now()
		}

		// Calculate length if not provided
		if blogPost.Length == 0 {
			blogPost.Length = len(blogPost.Content)
		}

		// Extract tags before creating the blog post
		tags := blogPost.Tags
		blogPost.Tags = nil // Clear tags to avoid issues during creation

		if err := h.blogPostRepo.Add(&blogPost); err != nil {
			h.responder.WriteError(w, wrapDatabaseError("create blog post", "blog_post", err))
			return
		}

		// Create tags if provided
		if len(tags) > 0 {
			for i := range tags {
				tags[i].BlogPostID = blogPost.ID
				if tags[i].ID == uuid.Nil {
					tags[i].ID = uuid.New()
				}
				if err := h.blogTagRepo.Add(&tags[i]); err != nil {
					h.logger.Error().Err(err).Str("tag_value", tags[i].Value).Msg("Failed to create blog tag")
					// Continue creating other tags even if one fails
				}
			}
		}

		// Reload blog post to get tags
		createdBlogPost, err := h.blogPostRepo.FindByID(blogPost.ID)
		if err != nil {
			h.responder.WriteError(w, wrapDatabaseError("find created blog post", "blog_post", err))
			return
		}

		// Get mainImageURL from query parameter (optional, for Substack posting)
		mainImageURL := r.URL.Query().Get("mainImageURL")

		// Get platforms to post to from query parameter (optional, comma-separated)
		// Valid values: substack, medium, twitter, linkedin
		// If not provided, defaults to all platforms for backward compatibility
		platformsParam := r.URL.Query().Get("platforms")
		var platformsToPost []string
		if platformsParam != "" {
			platformsToPost = strings.Split(platformsParam, ",")
			// Trim whitespace from each platform name
			for i := range platformsToPost {
				platformsToPost[i] = strings.TrimSpace(platformsToPost[i])
			}
			h.logger.Info().Strs("platforms", platformsToPost).Msg("Posting blog post to selected social media platforms")
		} else {
			// Default to all platforms for backward compatibility
			platformsToPost = []string{"substack", "medium", "twitter", "linkedin"}
			h.logger.Info().Msg("Posting blog post to all social media platforms")
		}

		if err := services.PostEverywhere(*createdBlogPost, createdBlogPost.Tags, mainImageURL, platformsToPost); err != nil {
			// Log the error but don't fail the request - the blog post was created successfully
			// The client can check logs or retry posting separately if needed
			h.logger.Error().Err(err).Msg("Failed to post to some social media platforms, but blog post was created successfully")
		}

		response := BlogPostWithTags{
			BlogPost: *createdBlogPost,
			Tags:     createdBlogPost.Tags,
		}

		w.WriteHeader(http.StatusCreated)
		h.responder.WriteJSON(w, response)
	}
}

// updateBlogPost updates an existing blog post
// @Summary Update blog post
// @Description Updates an existing blog post in the database
// @Tags Blog Posts
// @Accept json
// @Produce json
// @Param blogPostID path string true "Blog Post ID" format(uuid)
// @Param blogPost body models.BlogPost true "Updated blog post data"
// @Success 200 {object} BlogPostWithTags "Updated blog post with tags"
// @Failure 400 {object} errs.ErrorResponse "Bad Request - Invalid blog post data"
// @Failure 404 {object} errs.ErrorResponse "Not Found - Blog post not found"
// @Failure 500 {object} errs.ErrorResponse "Internal Server Error - Error updating blog post"
// @Router /blog-post/{blogPostID} [put]
func (h blogPostHandler) updateBlogPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authentication handled by middleware

		blogPostIDStr := chi.URLParam(r, "blogPostID")
		if blogPostIDStr == "" {
			h.responder.WriteError(w, errs.NewBadRequestError("missing blogPostID"))
			return
		}

		blogPostID, err := uuid.Parse(blogPostIDStr)
		if err != nil {
			h.responder.WriteError(w, errs.NewBadRequestError("invalid blogPostID"))
			return
		}

		// Verify blog post exists
		existingBlogPost, err := h.blogPostRepo.FindByID(blogPostID)
		if err != nil {
			h.responder.WriteError(w, wrapDatabaseError("find blog post", "blog_post", err))
			return
		}

		if existingBlogPost == nil {
			h.responder.WriteError(w, errs.NewNotFoundError("blog post not found"))
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to read request body")
			h.responder.WriteError(w, errs.NewBadRequestError("failed to read request body"))
			return
		}

		var blogPost models.BlogPost
		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&blogPost); err != nil {
			h.logger.Error().Err(err).Str("body", string(bodyBytes)).Msg("Failed to decode blog post request body")
			h.responder.WriteError(w, errs.NewBadRequestError("malformed request body"))
			return
		}

		// Ensure ID matches
		blogPost.ID = blogPostID

		// Update DateEdited
		now := time.Now()
		blogPost.DateEdited = &now

		// Update length if content changed
		if blogPost.Content != "" {
			blogPost.Length = len(blogPost.Content)
		}

		if err := h.blogPostRepo.Update(&blogPost); err != nil {
			h.responder.WriteError(w, wrapDatabaseError("update blog post", "blog_post", err))
			return
		}

		// Reload blog post to get updated tags
		updatedBlogPost, err := h.blogPostRepo.FindByID(blogPostID)
		if err != nil {
			h.responder.WriteError(w, wrapDatabaseError("find updated blog post", "blog_post", err))
			return
		}

		response := BlogPostWithTags{
			BlogPost: *updatedBlogPost,
			Tags:     updatedBlogPost.Tags,
		}

		h.responder.WriteJSON(w, response)
	}
}

// deleteBlogPost deletes a blog post by ID
// @Summary Delete blog post
// @Description Deletes a blog post from the database by ID
// @Tags Blog Posts
// @Accept json
// @Produce json
// @Param blogPostID path string true "Blog Post ID" format(uuid)
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} errs.ErrorResponse "Bad Request - Invalid blogPostID"
// @Failure 404 {object} errs.ErrorResponse "Not Found - Blog post not found"
// @Failure 500 {object} errs.ErrorResponse "Internal Server Error - Error deleting blog post"
// @Router /blog-post/{blogPostID} [delete]
func (h blogPostHandler) deleteBlogPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authentication handled by middleware

		blogPostIDStr := chi.URLParam(r, "blogPostID")
		if blogPostIDStr == "" {
			h.responder.WriteError(w, errs.NewBadRequestError("missing blogPostID"))
			return
		}

		blogPostID, err := uuid.Parse(blogPostIDStr)
		if err != nil {
			h.responder.WriteError(w, errs.NewBadRequestError("invalid blogPostID"))
			return
		}

		// Verify blog post exists
		_, err = h.blogPostRepo.FindByID(blogPostID)
		if err != nil {
			h.responder.WriteError(w, wrapDatabaseError("find blog post", "blog_post", err))
			return
		}

		if err := h.blogPostRepo.Delete(blogPostID); err != nil {
			h.responder.WriteError(w, wrapDatabaseError("delete blog post", "blog_post", err))
			return
		}

		h.responder.WriteJSON(w, map[string]string{
			"status":  "success",
			"message": "blog post deleted successfully",
		})
	}
}
