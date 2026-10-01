package presentation

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/smart-onestop/platform/internal/domain"
	"github.com/smart-onestop/platform/internal/infrastructure/config"
	"github.com/smart-onestop/platform/internal/infrastructure/security"
	"github.com/smart-onestop/platform/internal/presentation/handler"
	"github.com/smart-onestop/platform/internal/presentation/middleware"
)

type RouterConfig struct {
	Cfg              *config.Config
	JWTManager       *security.JWTManager
	AuthHandler      *handler.AuthHandler
	CaseHandler      *handler.CaseHandler
	HearingHandler   *handler.HearingHandler
	StructureHandler *handler.StructureHandler
	AnalyticsHandler *handler.AnalyticsHandler
}

// SetupRouter initializes the Gin router with all middleware, routes and static assets
func SetupRouter(rc *RouterConfig) *gin.Engine {
	if rc.Cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CorrelationID())
	r.Use(middleware.CORS(rc.Cfg.CORSAllowOrigins))

	idempotencyStore := middleware.NewIdempotencyStore()

	// 1. System Liveness & Observability
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "HEALTHY",
			"service":     "smart-onestop-platform",
			"environment": rc.Cfg.Environment,
		})
	})

	r.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, "# HELP http_requests_total Total number of HTTP requests\n# TYPE http_requests_total counter\nhttp_requests_total 1\n")
	})

	r.GET("/api/v1/docs", func(c *gin.Context) {
		spec, err := os.ReadFile("openapi.yaml")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "OpenAPI spec file not found"})
			return
		}
		c.Data(http.StatusOK, "text/yaml; charset=utf-8", spec)
	})

	// 2. Static Web Portal (Citizen & Official Unified Portal)
	if _, err := os.Stat("web/static"); err == nil {
		r.Static("/static", "web/static")
		r.StaticFile("/", "web/static/index.html")
	}

	// 3. API v1 Routes
	api := r.Group("/api/v1")
	{
		// --- Public Citizen & Municipal Hierarchy Endpoints ---
		api.GET("/structures/tree", rc.StructureHandler.GetTree)
		api.GET("/structures", rc.StructureHandler.ListAll)
		api.GET("/service-types", rc.StructureHandler.ListServiceTypes)

		// Public Citizen Grievance Intake with Idempotency protection
		api.POST("/cases", middleware.Idempotency(idempotencyStore), rc.CaseHandler.CreateCase)

		// Public Citizen Tracking & QR Verification
		api.GET("/cases/ticket/:ticket", rc.CaseHandler.GetByTicket)
		api.GET("/cases/verify/:qr", rc.CaseHandler.VerifyQR)

		// Official Authentication
		api.POST("/auth/login", rc.AuthHandler.Login)

		// --- Protected Official / Supervisor Endpoints ---
		authorized := api.Group("")
		authorized.Use(middleware.AuthRequired(rc.JWTManager))
		{
			authorized.GET("/auth/me", rc.AuthHandler.GetMe)

			// Cases Management
			authorized.GET("/cases", rc.CaseHandler.ListCases)
			authorized.GET("/cases/:id", rc.CaseHandler.GetByID)
			authorized.POST("/cases/:id/assign", rc.CaseHandler.AssignCase)
			authorized.POST("/cases/:id/attachments", rc.CaseHandler.AddAttachment)

			// Resolution (Woreda Officer, Sub-City Manager, City Director, Super Admin)
			authorized.POST("/cases/:id/resolve",
				middleware.RequireRoles(
					domain.RoleWoredaOfficer,
					domain.RoleSubcityManager,
					domain.RoleCityDirector,
					domain.RoleSuperAdmin,
				),
				rc.CaseHandler.ResolveCase,
			)

			// Manual Escalation (Sub-City Manager, City Director, Super Admin)
			authorized.POST("/cases/:id/escalate",
				middleware.RequireRoles(
					domain.RoleSubcityManager,
					domain.RoleCityDirector,
					domain.RoleSuperAdmin,
				),
				rc.CaseHandler.ManualEscalate,
			)

			// Hearing Sessions (Wednesday & Friday Desks)
			authorized.POST("/cases/:id/hearings", rc.HearingHandler.ScheduleSlot)
			authorized.GET("/hearings/date/:date", rc.HearingHandler.ListByDate)
			authorized.POST("/hearings/:id/complete", rc.HearingHandler.CompleteHearing)

			// Executive Analytics & SLA Operations
			authorized.GET("/analytics/dashboard", rc.AnalyticsHandler.GetDashboard)
			authorized.POST("/sla/trigger-sweep", rc.AnalyticsHandler.TriggerSLASweep)

			// Official Staff & Administrator User Management
			authorized.GET("/users", rc.AuthHandler.ListUsers)
			authorized.POST("/users", rc.AuthHandler.CreateUser)
			authorized.DELETE("/users/:id", rc.AuthHandler.DeleteUser)
		}
	}

	return r
}
