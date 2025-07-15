package http

import (
	
	"vybes/internal/config"
	"vybes/internal/middleware"
	"vybes/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes the Gin router and sets up the routes.
func SetupRouter(
	userHandler *UserHandler,
	followHandler *FollowHandler,
	suggestionHandler *SuggestionHandler,
	storyHandler *StoryHandler,
	contentHandler *ContentHandler,
	reactionHandler *ReactionHandler,
	feedHandler *FeedHandler,
	bookmarkHandler *BookmarkHandler,
	searchHandler *SearchHandler,
	notificationHandler *NotificationHandler,
	sessionHandler *SessionHandler,
	sessionService *service.SessionService,
	cfg *config.Config,
) *gin.Engine {
	router := gin.Default()

	// API v1 routes
	apiV1 := router.Group("/api/v1")
	{
		// Serve static files from the 'public' directory
		// e.g., /api/v1/llm.txt will serve the public/llm.txt file
		apiV1.StaticFS("/public", gin.Dir("./public", false))

		// Public routes
		publicUserRoutes := apiV1.Group("/users")
		{
			publicUserRoutes.POST("/register", userHandler.Register)
			publicUserRoutes.POST("/login", userHandler.Login)
			publicUserRoutes.POST("/refresh", userHandler.RefreshToken)
			publicUserRoutes.POST("/request-otp", userHandler.RequestOTP)
			publicUserRoutes.POST("/reset-password", userHandler.ResetPassword)
		}

		publicPostRoutes := apiV1.Group("/posts")
		{
			publicPostRoutes.POST("/:postID/view", contentHandler.RecordView)
		}

		// Authenticated routes
		authRoutes := apiV1.Group("/")
		authRoutes.Use(middleware.AuthMiddleware(cfg, sessionService))
		{
			// Search routes
			authRoutes.GET("/search/users", searchHandler.SearchUsers)

			// User profile and wallet routes
			authRoutes.GET("/users/:username", userHandler.GetUserProfile)
			authRoutes.PATCH("/users/me", userHandler.UpdateProfile)
			authRoutes.POST("/wallet/unlock", userHandler.UnlockWallet)
			authRoutes.POST("/wallet/export", userHandler.ExportPrivateKey)
			authRoutes.POST("/wallet/personal-sign", userHandler.PersonalSign)
			authRoutes.POST("/wallet/sign-transaction", userHandler.SignTransaction)
			authRoutes.POST("/wallet/send-transaction", userHandler.SendTransaction)
			authRoutes.POST("/wallet/sign-typed-data", userHandler.SignTypedDataV4)
			authRoutes.POST("/wallet/secp256k1-sign", userHandler.Secp256k1Sign)

			// Follow routes
			authRoutes.POST("/users/:username/follow", followHandler.FollowUser)
			authRoutes.DELETE("/users/:username/follow", followHandler.UnfollowUser)

			// Suggestion routes
			authRoutes.GET("/suggestions/users", suggestionHandler.GetSuggestions)

			// Story routes
			authRoutes.POST("/stories", storyHandler.CreateStory)
			authRoutes.GET("/stories/feed", storyHandler.GetStoryFeed)

			// Post and Content routes
			posts := authRoutes.Group("/posts")
			{
				posts.POST("/", contentHandler.CreatePost)
				posts.DELETE("/:postID", contentHandler.DeletePost)
				posts.POST("/:postID/repost", contentHandler.Repost)
				posts.GET("/:postID/comments", contentHandler.GetComments)
				posts.POST("/:postID/comments", contentHandler.CreateComment)
				posts.POST("/:postID/like", reactionHandler.AddLike)
				posts.DELETE("/:postID/like", reactionHandler.RemoveLike)
				posts.POST("/:postID/bookmark", bookmarkHandler.AddBookmark)
				posts.DELETE("/:postID/bookmark", bookmarkHandler.RemoveBookmark)
			}
			
			// Repost-specific routes
			authRoutes.GET("/reposts/by-user/:userID", contentHandler.GetRepostsByUser)

			// Feed and Bookmark routes
			authRoutes.GET("/feeds/for-you", feedHandler.GetForYouFeed)
			authRoutes.GET("/feeds/friends", feedHandler.GetFriendFeed)
			authRoutes.GET("/bookmarks", bookmarkHandler.GetBookmarks)

			// Notification routes
			notifications := authRoutes.Group("/notifications")
			{
				notifications.GET("/", notificationHandler.GetNotifications)
				notifications.PATCH("/read", notificationHandler.MarkAsRead)
			}

			// Session routes
			sessions := authRoutes.Group("/sessions")
			{
				sessions.GET("/:id", sessionHandler.GetSession)
				sessions.POST("/:id/block", sessionHandler.BlockSession)
			}
		}
	}

	return router
}