package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/ProNexus-Startup/ProNexus/backend/database"
	"github.com/ProNexus-Startup/ProNexus/backend/errs"
	"github.com/ProNexus-Startup/ProNexus/backend/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type projectHandler struct {
	responder      Responder
	logger         zerolog.Logger
	projectRepo    *database.ProjectRepo
	projectTagRepo *database.ProjectTagRepo
}

func newProjectHandler(projectRepo *database.ProjectRepo, projectTagRepo *database.ProjectTagRepo) projectHandler {
	logger := log.With().Str("handlerName", "projectHandler").Logger()

	return projectHandler{
		responder:      NewResponder(logger),
		logger:         logger,
		projectRepo:    projectRepo,
		projectTagRepo: projectTagRepo,
	}
}

// ProjectWithTags represents a project with its tags
type ProjectWithTags struct {
	Project models.Project      `json:"project"`
	Tags    []models.ProjectTag `json:"tags"`
}

// ProjectCollectionWithTags represents multiple projects with their tags
type ProjectCollectionWithTags struct {
	Projects []ProjectWithTags `json:"projects"`
	Total    int               `json:"total,omitempty"`
}

// getAllProjects retrieves all projects with their tags
// @Summary Get all projects
// @Description Retrieves all projects from the database with their associated tags
// @Tags Projects
// @Accept json
// @Produce json
// @Success 200 {object} ProjectCollectionWithTags "List of projects with tags"
// @Failure 500 {object} errs.ErrorResponse "Internal Server Error - Error fetching projects"
// @Router /projects [get]
func (h projectHandler) getAllProjects() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authentication handled by middleware

		projects, err := h.projectRepo.FindAll()
		if err != nil {
			h.responder.WriteError(w, wrapDatabaseError("find projects", "projects", err))
			return
		}

		// Convert to ProjectWithTags format
		var projectsWithTags []ProjectWithTags
		for _, project := range projects {
			projectsWithTags = append(projectsWithTags, ProjectWithTags{
				Project: *project,
				Tags:    project.Tags,
			})
		}

		response := ProjectCollectionWithTags{
			Projects: projectsWithTags,
			Total:    len(projectsWithTags),
		}

		h.responder.WriteJSON(w, response)
	}
}

// getProject retrieves a specific project by ID with its tags
// @Summary Get project
// @Description Retrieves detailed information about a specific project by ID with its tags
// @Tags Projects
// @Accept json
// @Produce json
// @Param projectID path string true "Project ID" format(uuid)
// @Success 200 {object} ProjectWithTags "Project details with tags"
// @Failure 400 {object} errs.ErrorResponse "Bad Request - Invalid projectID"
// @Failure 404 {object} errs.ErrorResponse "Not Found - Project not found"
// @Failure 500 {object} errs.ErrorResponse "Internal Server Error - Error fetching project"
// @Router /project/{projectID} [get]
func (h projectHandler) getProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authentication handled by middleware

		projectIDStr := chi.URLParam(r, "projectID")
		if projectIDStr == "" {
			h.responder.WriteError(w, errs.NewBadRequestError("missing projectID"))
			return
		}

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			h.responder.WriteError(w, errs.NewBadRequestError("invalid projectID"))
			return
		}

		project, err := h.projectRepo.FindByID(projectID)
		if err != nil {
			h.responder.WriteError(w, wrapDatabaseError("find project", "project", err))
			return
		}

		if project == nil {
			h.responder.WriteError(w, errs.NewNotFoundError("project not found"))
			return
		}

		response := ProjectWithTags{
			Project: *project,
			Tags:    project.Tags,
		}

		h.responder.WriteJSON(w, response)
	}
}

// createProject creates a new project
// @Summary Create project
// @Description Creates a new project in the database
// @Tags Projects
// @Accept json
// @Produce json
// @Param project body models.Project true "Project data"
// @Success 201 {object} ProjectWithTags "Created project with tags"
// @Failure 400 {object} errs.ErrorResponse "Bad Request - Invalid project data"
// @Failure 500 {object} errs.ErrorResponse "Internal Server Error - Error creating project"
// @Router /project [post]
func (h projectHandler) createProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authentication handled by middleware

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to read request body")
			h.responder.WriteError(w, errs.NewBadRequestError("failed to read request body"))
			return
		}

		var project models.Project
		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&project); err != nil {
			h.logger.Error().Err(err).Str("body", string(bodyBytes)).Msg("Failed to decode project request body")
			h.responder.WriteError(w, errs.NewBadRequestError("malformed request body"))
			return
		}

		if project.Title == "" {
			h.responder.WriteError(w, errs.NewBadRequestError("title is required"))
			return
		}

		// Extract tags before creating the project
		tags := project.Tags
		project.Tags = nil // Clear tags to avoid issues during creation

		if err := h.projectRepo.Add(&project); err != nil {
			h.responder.WriteError(w, wrapDatabaseError("create project", "project", err))
			return
		}

		// Create tags if provided
		if len(tags) > 0 {
			for i := range tags {
				tags[i].ProjectID = project.ID
				if tags[i].ID == uuid.Nil {
					tags[i].ID = uuid.New()
				}
				if err := h.projectTagRepo.Add(&tags[i]); err != nil {
					h.logger.Error().Err(err).Str("tag_value", tags[i].Value).Msg("Failed to create project tag")
					// Continue creating other tags even if one fails
				}
			}
		}

		// Reload project to get tags
		createdProject, err := h.projectRepo.FindByID(project.ID)
		if err != nil {
			h.responder.WriteError(w, wrapDatabaseError("find created project", "project", err))
			return
		}

		response := ProjectWithTags{
			Project: *createdProject,
			Tags:    createdProject.Tags,
		}

		w.WriteHeader(http.StatusCreated)
		h.responder.WriteJSON(w, response)
	}
}

// updateProject updates an existing project
// @Summary Update project
// @Description Updates an existing project in the database
// @Tags Projects
// @Accept json
// @Produce json
// @Param projectID path string true "Project ID" format(uuid)
// @Param project body models.Project true "Updated project data"
// @Success 200 {object} ProjectWithTags "Updated project with tags"
// @Failure 400 {object} errs.ErrorResponse "Bad Request - Invalid project data"
// @Failure 404 {object} errs.ErrorResponse "Not Found - Project not found"
// @Failure 500 {object} errs.ErrorResponse "Internal Server Error - Error updating project"
// @Router /project/{projectID} [put]
func (h projectHandler) updateProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authentication handled by middleware

		projectIDStr := chi.URLParam(r, "projectID")
		if projectIDStr == "" {
			h.responder.WriteError(w, errs.NewBadRequestError("missing projectID"))
			return
		}

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			h.responder.WriteError(w, errs.NewBadRequestError("invalid projectID"))
			return
		}

		// Verify project exists
		existingProject, err := h.projectRepo.FindByID(projectID)
		if err != nil {
			h.responder.WriteError(w, wrapDatabaseError("find project", "project", err))
			return
		}

		if existingProject == nil {
			h.responder.WriteError(w, errs.NewNotFoundError("project not found"))
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to read request body")
			h.responder.WriteError(w, errs.NewBadRequestError("failed to read request body"))
			return
		}

		var project models.Project
		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&project); err != nil {
			h.logger.Error().Err(err).Str("body", string(bodyBytes)).Msg("Failed to decode project request body")
			h.responder.WriteError(w, errs.NewBadRequestError("malformed request body"))
			return
		}

		// Ensure ID matches
		project.ID = projectID

		if err := h.projectRepo.Update(&project); err != nil {
			h.responder.WriteError(w, wrapDatabaseError("update project", "project", err))
			return
		}

		// Reload project to get updated tags
		updatedProject, err := h.projectRepo.FindByID(projectID)
		if err != nil {
			h.responder.WriteError(w, wrapDatabaseError("find updated project", "project", err))
			return
		}

		response := ProjectWithTags{
			Project: *updatedProject,
			Tags:    updatedProject.Tags,
		}

		h.responder.WriteJSON(w, response)
	}
}

// deleteProject deletes a project by ID
// @Summary Delete project
// @Description Deletes a project from the database by ID
// @Tags Projects
// @Accept json
// @Produce json
// @Param projectID path string true "Project ID" format(uuid)
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} errs.ErrorResponse "Bad Request - Invalid projectID"
// @Failure 404 {object} errs.ErrorResponse "Not Found - Project not found"
// @Failure 500 {object} errs.ErrorResponse "Internal Server Error - Error deleting project"
// @Router /project/{projectID} [delete]
func (h projectHandler) deleteProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authentication handled by middleware

		projectIDStr := chi.URLParam(r, "projectID")
		if projectIDStr == "" {
			h.responder.WriteError(w, errs.NewBadRequestError("missing projectID"))
			return
		}

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			h.responder.WriteError(w, errs.NewBadRequestError("invalid projectID"))
			return
		}

		// Verify project exists
		_, err = h.projectRepo.FindByID(projectID)
		if err != nil {
			h.responder.WriteError(w, wrapDatabaseError("find project", "project", err))
			return
		}

		if err := h.projectRepo.Delete(projectID); err != nil {
			h.responder.WriteError(w, wrapDatabaseError("delete project", "project", err))
			return
		}

		h.responder.WriteJSON(w, map[string]string{
			"status":  "success",
			"message": "project deleted successfully",
		})
	}
}
