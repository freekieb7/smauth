package web

import (
	"context"
	"errors"

	"github.com/freekieb7/smauth/internal/http"
	"github.com/freekieb7/smauth/internal/openehr"
	"github.com/freekieb7/smauth/internal/openehr/aql"
	"github.com/freekieb7/smauth/internal/telemetry"
)

type OpenEHRHandler struct {
	Logger     *telemetry.Logger
	Store      *openehr.Store
	AQLService *aql.Service
	Validator  *openehr.Validator
}

func NewOpenEHRHandler(logger *telemetry.Logger, store *openehr.Store, aqlService *aql.Service, validator *openehr.Validator) OpenEHRHandler {
	return OpenEHRHandler{
		Logger:     logger,
		Store:      store,
		AQLService: aqlService,
		Validator:  validator,
	}
}

func (h *OpenEHRHandler) RegisterRoutes(router *http.Router) {
	// OpenEHR API routes
	router.Group("/openehr/v1", func(group *http.Router) {
		group.GET("/", h.ServerInfo)

		group.Group("/ehrs", func(group *http.Router) {
			group.GET("", h.ListEHRs)
			group.POST("", h.CreateEHR)

			group.Group("/{ehr_id}", func(group *http.Router) {
				group.GET("", h.GetEHRByID)
				group.DELETE("", h.DeleteEHRByID)

				group.Group("/ehr_status", func(group *http.Router) {
					group.GET("", h.GetEHRStatusByID)
					group.PUT("", h.UpdateEHRStatusByID)
				})

				group.Group("/compositions", func(group *http.Router) {
					group.GET("", h.ListCompositions)
					group.POST("", h.CreateComposition)

					group.Group("/{composition_id}", func(group *http.Router) {
						group.GET("", h.GetCompositionByID)
						group.PUT("", h.UpdateCompositionByID)
						group.DELETE("", h.DeleteCompositionByID)
					})
				})

				group.Group("/folder", func(group *http.Router) {
					group.GET("", h.ListFolders)
					group.POST("", h.CreateFolder)

					group.Group("/{folder_id}", func(group *http.Router) {
						group.GET("", h.GetFolderByID)
						group.PUT("", h.UpdateFolderByID)
						group.DELETE("", h.DeleteFolderByID)
					})
				})

				// todo contribution (just not sure what the difference is with demographics)
			})
		})

		group.Group("/templates", func(group *http.Router) {
			group.GET("", h.ListTemplates)
			group.POST("", h.CreateTemplate)

			group.Group("/{template_id}", func(group *http.Router) {
				group.GET("", h.GetTemplateByID)
				group.PUT("", h.UpdateTemplateByID)
				group.DELETE("", h.DeleteTemplateByID)
			})
		})

		group.Group("/agents", func(group *http.Router) {
			group.GET("", h.ListAgent)
			group.POST("", h.CreateAgent)

			group.Group("/{agent_id}", func(group *http.Router) {
				group.GET("", h.GetAgentByID)
				group.PUT("", h.UpdateAgentByID)
				group.DELETE("", h.DeleteAgentByID)
			})
		})

		group.Group("/groups", func(group *http.Router) {
			group.GET("", h.ListGroup)
			group.POST("", h.CreateGroup)

			group.Group("/{group_id}", func(group *http.Router) {
				group.GET("", h.GetGroupByID)
				group.PUT("", h.UpdateGroupByID)
				group.DELETE("", h.DeleteGroupByID)
			})
		})

		group.Group("/organisations", func(group *http.Router) {
			group.GET("", h.ListOrganisation)
			group.POST("", h.CreateOrganisation)

			group.Group("/{organisation_id}", func(group *http.Router) {
				group.GET("", h.GetOrganisationByID)
				group.PUT("", h.UpdateOrganisationByID)
				group.DELETE("", h.DeleteOrganisationByID)
			})
		})

		group.Group("/persons", func(group *http.Router) {
			group.GET("", h.ListPerson)
			group.POST("", h.CreatePerson)

			group.Group("/{person_id}", func(group *http.Router) {
				group.GET("", h.GetPersonByID)
				group.PUT("", h.UpdatePersonByID)
				group.DELETE("", h.DeletePersonByID)
			})
		})

		group.Group("/roles", func(group *http.Router) {
			group.GET("", h.ListRole)
			group.POST("", h.CreateRole)

			group.Group("/{role_id}", func(group *http.Router) {
				group.GET("", h.GetRoleByID)
				group.PUT("", h.UpdateRoleByID)
				group.DELETE("", h.DeleteRoleByID)
			})
		})

		group.Group("/query", func(group *http.Router) {
			group.POST("", h.ExecuteQuery)
			group.POST("/prepared", h.CreatePreparedTable)
			group.POST("/prepared/{name}/sync", h.SyncPreparedTable)
		})
	})
}

func (h *OpenEHRHandler) ServerInfo(ctx context.Context, req *http.Request, res *http.Response) error {
	return res.SendJSON(http.StatusOK, map[string]any{
		"solution":              "openEHRSys",
		"solution_version":      "v1.0",
		"vendor":                "GOpenEHR",
		"restapi_specs_version": "1.0.3",
		"conformance_profile":   "CUSTOM",
		"endpoints": []string{
			"/ehr",
			"/demographics",
			"/definition",
			"/query",
			"/admin",
		},
	})
}

func (h *OpenEHRHandler) ListEHRs(ctx context.Context, req *http.Request, res *http.Response) error {
	query := `SELECT e FROM EHR e LIMIT 100`

	res.SetHeader("content-type", "application/json")

	_, err := h.AQLService.Query(ctx, res.Writer, query, map[string]any{}, []aql.PreparedTable{})
	if err != nil {
		h.Logger.Error("AQL query failed", "error", err)
		return err
	}

	return nil
}

func (h *OpenEHRHandler) CreateEHR(ctx context.Context, req *http.Request, res *http.Response) error {
	ehr := h.Store.NewEHR()

	if errs := h.Validator.ValidateModel(ehr); errs != nil {
		h.Logger.Error("EHR validation failed", "errors", errs)
		return res.Send(http.StatusInternalServerError)
	}

	err := h.Store.SaveEHR(ctx, ehr)
	if err != nil {
		h.Logger.Error("Failed to save EHR", "error", err)
		return err
	}

	h.Logger.Info("EHR created successfully", "ehr_id", ehr.EHRID.Value)

	return res.SendJSON(http.StatusCreated, ehr)
}

func (h *OpenEHRHandler) GetEHRByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) DeleteEHRByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) GetEHRStatusByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) UpdateEHRStatusByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) ListCompositions(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) CreateComposition(ctx context.Context, req *http.Request, res *http.Response) error {
	var composition openehr.COMPOSITION
	if err := req.DecodeJSON(&composition); err != nil {
		return err
	}

	if errs := h.Validator.ValidateModel(composition); errs != nil {
		h.Logger.Error("Composition validation failed", "errors", errs)
		return res.SendJSON(http.StatusBadRequest, errs)
	}

	return nil
}

func (h *OpenEHRHandler) GetCompositionByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) UpdateCompositionByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) DeleteCompositionByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) ListFolders(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) CreateFolder(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) GetFolderByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) UpdateFolderByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) DeleteFolderByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) ListTemplates(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) CreateTemplate(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) GetTemplateByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) UpdateTemplateByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) DeleteTemplateByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) ListAgent(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) CreateAgent(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) GetAgentByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) UpdateAgentByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) DeleteAgentByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) ListGroup(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) CreateGroup(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) GetGroupByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) UpdateGroupByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) DeleteGroupByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) ListOrganisation(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) CreateOrganisation(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) GetOrganisationByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) UpdateOrganisationByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) DeleteOrganisationByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) ListPerson(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) CreatePerson(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) GetPersonByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) UpdatePersonByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) DeletePersonByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) ListRole(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) CreateRole(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) GetRoleByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) UpdateRoleByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) DeleteRoleByID(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) ExecuteQuery(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) CreatePreparedTable(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}

func (h *OpenEHRHandler) SyncPreparedTable(ctx context.Context, req *http.Request, res *http.Response) error {
	return errors.New("not implemented")
}
