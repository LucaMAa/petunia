package router

import (
	"encoding/json"
	"log"
	"petunia/internal/config"
	"petunia/internal/controller"
	"petunia/internal/cron"
	"petunia/internal/middleware"
	"petunia/internal/model"
	"petunia/internal/repository"
	"petunia/internal/service"
	"petunia/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Setup() *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	// ── Repositories ──────────────────────────────────────────────────────────
	refreshRepo := repository.NewRefreshTokenRepository()
	resetRepo := repository.NewPasswordResetRepository()
	userRepo := repository.NewUserRepository()
	emailChangeRepo := repository.NewEmailChangeRepository()
	petRepo := repository.NewPetRepository()
	familyRepo := repository.NewFamilyRepository()
	inviteRepo := repository.NewFamilyInviteRepository()
	reminderRepo := repository.NewReminderRepository()
	firedLogRepo := repository.NewReminderFiredLogRepository()
	fileRepo := repository.NewUploadedFileRepository()
	pushTokenRepo := repository.NewPushTokenRepository()

	var geoSvc service.GeoService = service.NewGeoService(config.Redis)
	mapReportRepo := repository.NewMapReportRepository(geoSvc)
	activityRepo := repository.NewActivityRepository()

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, refreshRepo, resetRepo)
	profileSvc := service.NewProfileService(userRepo, emailChangeRepo)
	petSvc := service.NewPetService(petRepo, userRepo)
	pushSvc := service.NewPushService(pushTokenRepo)
	familySvc := service.NewFamilyService(familyRepo, userRepo, petRepo, inviteRepo, pushSvc)
	mapReportSvc := service.NewMapReportService(mapReportRepo, userRepo, geoSvc, pushSvc)
	activitySvc := service.NewActivityService(activityRepo, petRepo)
	reminderSvc := service.NewReminderService(reminderRepo, familyRepo, userRepo, firedLogRepo, pushSvc)
	uploadSvc := service.NewUploadedFileService(fileRepo, petRepo, userRepo, familyRepo)

	// ── Controllers ───────────────────────────────────────────────────────────
	authCtrl := controller.NewAuthController(authSvc)
	profileCtrl := controller.NewProfileController(profileSvc)
	petCtrl := controller.NewPetController(petSvc)
	familyCtrl := controller.NewFamilyController(familySvc)
	mapReportCtrl := controller.NewMapReportController(mapReportSvc)
	activityCtrl := controller.NewActivityController(activitySvc)
	reminderCtrl := controller.NewReminderController(reminderSvc)
	uploadCtrl := controller.NewUploadController(uploadSvc, fileRepo)
	pushCtrl := controller.NewPushController(pushTokenRepo)
	locationCtrl := controller.NewLocationController(geoSvc)

	// ── Cron factory ─────────────────────────────────────────────────────────
	cronCfg := cron.ConfigFromFile("cron_config.json")
	cronFactory := cron.NewFactory(cronCfg, reminderRepo, reminderSvc)
	cronFactory.Start()

	// ── WebSocket ─────────────────────────────────────────────────────────────
	wsMessageHandler := func(userID uuid.UUID, msg ws.IncomingMessage) {
		switch msg.Event {
		case ws.EventLocationUpdate:
			var loc ws.LocationUpdatePayload
			if err := json.Unmarshal(msg.Payload, &loc); err == nil {
				ws.GlobalHub.UpdateLocation(userID, loc.Lat, loc.Lng)
				geoSvc.SetUserLocation(userID.String(), loc.Lat, loc.Lng)
			}
		default:
			log.Printf("[ws] message from %s: event=%s", userID, msg.Event)
		}
	}
	r.GET("/ws", ws.WsHandler(wsMessageHandler))

	// ── Routes ────────────────────────────────────────────────────────────────
	v1 := r.Group("/api")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authCtrl.Register)
			auth.POST("/login", authCtrl.Login)
			auth.POST("/refresh", authCtrl.Refresh)
			auth.POST("/logout", authCtrl.Logout)
			auth.POST("/google", authCtrl.Google)
			auth.POST("/change-password", middleware.Auth(), authCtrl.ChangePassword)
			auth.POST("/request-reset", authCtrl.RequestPasswordReset)
			auth.POST("/reset-password", authCtrl.ResetPassword)
		}

		profile := v1.Group("/profile", middleware.Auth())
		{
			profile.GET("", profileCtrl.GetProfile)
			profile.PATCH("", profileCtrl.UpdateProfile)
			profile.POST("/request-email-change", profileCtrl.RequestEmailChange)
			profile.POST("/change-password", profileCtrl.ChangePassword)
			profile.POST("/disable", profileCtrl.DisableAccount)
		}

		v1.POST("/confirm-email", profileCtrl.ConfirmEmailChange)
		v1.POST("/location", middleware.Auth(), locationCtrl.UpdateLocation)

		pets := v1.Group("/pets", middleware.Auth())
		{
			pets.POST("", petCtrl.CreatePet)
			pets.GET("", petCtrl.GetMyPets)
			pets.GET("/:id", petCtrl.GetPet)
			pets.PATCH("/:id", petCtrl.UpdatePet)
			pets.DELETE("/:id", petCtrl.DeletePet)
			pets.POST("/:id/owners", petCtrl.AddOwner)
			pets.DELETE("/:id/owners/:user_id", petCtrl.RemoveOwner)
		}

		families := v1.Group("/families", middleware.Auth())
		{
			families.POST("", familyCtrl.CreateFamily)
			families.GET("", familyCtrl.GetMyFamilies)
			families.GET("/:id", familyCtrl.GetFamily)
			families.PATCH("/:id", familyCtrl.UpdateFamily)
			families.DELETE("/:id", familyCtrl.DeleteFamily)
			families.POST("/:id/members", familyCtrl.SendInvite)
			families.DELETE("/:id/members/:user_id", familyCtrl.RemoveMember)
			families.POST("/:id/leave", familyCtrl.LeaveFamily)
			families.POST("/:id/pets", familyCtrl.AssignPet)
			families.DELETE("/:id/pets/:pet_id", familyCtrl.UnassignPet)
			families.GET("/search/users", familyCtrl.SearchUsers)
			families.GET("/:id/reminders", reminderCtrl.GetByFamily)
			families.GET("/:id/reminders/acks", reminderCtrl.GetAckHistory)
		}

		invites := v1.Group("/invites", middleware.Auth())
		{
			invites.GET("", familyCtrl.GetPendingInvites)
			invites.POST("/respond", familyCtrl.RespondToInvite)
			invites.GET("/sent", familyCtrl.GetSentInvites)
			invites.DELETE("", familyCtrl.CancelInvite)
		}

		reports := v1.Group("/reports", middleware.Auth())
		{
			reports.POST("", mapReportCtrl.CreateReport)
			reports.GET("/nearby", mapReportCtrl.GetNearby)
			reports.GET("/mine", mapReportCtrl.GetMyReports)
			reports.GET("/:id", mapReportCtrl.GetReport)
			reports.DELETE("/:id", mapReportCtrl.DeleteReport)
			reports.POST("/:id/abuse", mapReportCtrl.ReportAbuse)
		}

		activities := v1.Group("/activities", middleware.Auth())
		{
			activities.POST("", activityCtrl.Start)
			activities.GET("", activityCtrl.List)
			activities.GET("/:id", activityCtrl.Get)
			activities.POST("/:id/points", activityCtrl.Points)
			activities.POST("/:id/pause", activityCtrl.Status(model.ActivityStatusPaused))
			activities.POST("/:id/resume", activityCtrl.Status(model.ActivityStatusActive))
			activities.POST("/:id/finish", activityCtrl.Status(model.ActivityStatusCompleted))
			activities.DELETE("/:id", activityCtrl.Cancel)
		}

		reminders := v1.Group("/reminders", middleware.Auth())
		{
			reminders.POST("", reminderCtrl.Create)
			reminders.PATCH("/:id", reminderCtrl.Update)
			reminders.DELETE("/:id", reminderCtrl.Delete)
			reminders.POST("/:id/ack", reminderCtrl.Ack)
			reminders.GET("/pending-alerts", reminderCtrl.GetPendingAlerts)
		}

		// Translations
		transSvc := service.NewTranslationService(config.Redis)
		transCtrl := controller.NewTranslationController(transSvc)
		v1.GET("/translations", transCtrl.GetTranslations)

		upload := v1.Group("/upload", middleware.Auth())
		{
			upload.POST("/avatar/pets/:id", uploadCtrl.UploadPetAvatar)
			upload.POST("/avatar/users/me", uploadCtrl.UploadUserAvatar)
			upload.POST("/avatar/families/:id", uploadCtrl.UploadFamilyAvatar)
			upload.POST("/documents/pets/:id", uploadCtrl.UploadPetDocument)
			upload.GET("/documents/pets/:id", uploadCtrl.GetPetDocuments)
			upload.DELETE("/:id", uploadCtrl.DeleteFile)
		}

		filesGroup := v1.Group("/files")
		filesGroup.GET(":id", uploadCtrl.ServeFile)
		filesGroup.GET(":id/stream", uploadCtrl.ServeFile)

		pushGroup := v1.Group("/push-tokens", middleware.Auth())
		{
			pushGroup.POST("", pushCtrl.Register)
			pushGroup.DELETE("", pushCtrl.Unregister)
		}
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
