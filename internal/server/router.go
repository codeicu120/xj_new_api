package server

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"xj_comp/internal/config"
	"xj_comp/internal/handler"
	amazingRepo "xj_comp/internal/repository/amazing"
	commentRepo "xj_comp/internal/repository/comment"
	gameRepo "xj_comp/internal/repository/game"
	onegoRepo "xj_comp/internal/repository/onego"
	soRepo "xj_comp/internal/repository/so"
	ucpRepo "xj_comp/internal/repository/ucp"
	userRepo "xj_comp/internal/repository/user"
	vodRepo "xj_comp/internal/repository/vod"
	amazingService "xj_comp/internal/service/amazing"
	captchaService "xj_comp/internal/service/captcha"
	commentService "xj_comp/internal/service/comment"
	gameService "xj_comp/internal/service/game"
	iplocService "xj_comp/internal/service/iploc"
	onegoService "xj_comp/internal/service/onego"
	sendfileService "xj_comp/internal/service/sendfile"
	soService "xj_comp/internal/service/so"
	ucpService "xj_comp/internal/service/ucp"
	userService "xj_comp/internal/service/user"
	vodService "xj_comp/internal/service/vod"
)

type Options struct {
	Config config.Config
	Logger *slog.Logger
}

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Env     string `json:"env"`
	Time    string `json:"time"`
}

func NewRouter(opts Options) *gin.Engine {
	cfg := opts.Config
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	if cfg.Env == "prod" || cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.TestMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger(logger))

	userHandler := handler.NewUserHandler(userService.NewSysAvatarService(cfg.ResourceBaseURL))
	captchaHandler := handler.NewCaptchaHandler(captchaService.NewService(cfg.SMSCaptcha, cfg.CaptchaStyle, nil))
	ipLocHandler := handler.NewIPLocHandler(iplocService.NewService(newIPLocator(cfg.IPDBPath, logger)))
	db := openMySQL(cfg.MySQLDSN, logger)
	gameHandler := handler.NewGameHandler(
		gameService.NewPlatformService(gameRepo.NewPlatformRepository(db)),
		gameService.NewCategoryService(gameRepo.NewCategoryRepository(db), cfg.GameResourceURL),
		gameService.NewListingService(gameRepo.NewGameRepository(db), cfg.ResourceBaseURL),
		gameService.NewBroadcastService(gameRepo.NewBroadcastRepository(db)),
	)
	amazingHandler := handler.NewAmazingHandler(
		amazingService.NewCategoryService(amazingRepo.NewCategoryRepository(db)),
		amazingService.NewListingService(amazingRepo.NewSoftwareRepository(db), cfg.ResourceBaseURL),
	)
	soHandler := handler.NewSOHandler(soService.NewConfigService(soRepo.NewConfigRepository(db)))
	vodHandler := handler.NewVODHandler(vodService.NewListingService(vodRepo.NewListingRepository(db), cfg.ResourceBaseURL, cfg.VIPDiscount))
	userRepository := userRepo.NewRepository(db)
	ucpHandler := handler.NewUCPHandler(ucpService.NewService(ucpStore{user: userRepository, ucp: ucpRepo.NewRepository(db)}, cfg.ResourceBaseURL))
	sendfileHandler := handler.NewSendfileHandler(sendfileService.NewService(userRepository, vodRepo.NewListingRepository(db)))
	commentHandler := handler.NewCommentHandler(commentService.NewService(commentRepo.NewRepository(db), cfg.ResourceBaseURL))
	onegoHandler := handler.NewOneGoHandler(onegoService.NewService(onegoRepo.NewRepository(db)))

	router.GET("/healthz", healthHandler(cfg))
	router.GET("/readyz", healthHandler(cfg))
	router.Any("/sysavatar", userHandler.SysAvatar)
	router.Any("/captcha/req", captchaHandler.Req)
	router.Any("/iploc/:ip", ipLocHandler.Find)
	router.Any("/game/platforms", gameHandler.Platforms)
	router.Any("/game/categories", gameHandler.Categories)
	router.Any("/game/games", gameHandler.Games)
	router.Any("/game/broadcasts", gameHandler.Broadcasts)
	router.Any("/game/wali/gameList", gameHandler.WaliGames)
	router.Any("/getLikeRows", vodHandler.LikeRows)
	router.Any("/ucp/index", ucpHandler.Index)
	router.GET("/ucp/feedback", ucpHandler.FeedbackListing)
	router.GET("/ucp/feedback/index", ucpHandler.FeedbackIndex)
	router.GET("/ucp/feedback/listing", ucpHandler.FeedbackNewListing)
	router.GET("/ucp/feedback/detail", ucpHandler.FeedbackDetail)
	router.GET("/ucp/msg", ucpHandler.MsgListing)
	router.GET("/ucp/msg/index", ucpHandler.MsgListing)
	router.Any("/ucp/myaff", ucpHandler.MyAff)
	router.Any("/ucp/rolltitle", ucpHandler.RollTitle)
	router.Any("/ucp/affcenter", ucpHandler.AffCenter)
	router.Any("/ucp/payment", ucpHandler.PaymentListing)
	router.Any("/ucp/payment/index", ucpHandler.PaymentListing)
	router.Any("/ucp/payment/listing", ucpHandler.PaymentListing)
	router.Any("/ucp/payment/safepaylog", ucpHandler.SafePayLog)
	router.Any("/ucp/account", ucpHandler.AccountIndex)
	router.Any("/ucp/account/index", ucpHandler.AccountIndex)
	router.Any("/ucp/account/balancelog", ucpHandler.BalanceLog)
	router.Any("/ucp/coinlog", ucpHandler.CoinLogIndex)
	router.Any("/ucp/coinlog/index", ucpHandler.CoinLogIndex)
	router.Any("/ucp/coinlog/bonuslog", ucpHandler.CoinLogBonusLog)
	router.Any("/ucp/coinlog/invitelog", ucpHandler.CoinLogInviteLog)
	router.Any("/vod/show/:vodid", vodHandler.Show)
	router.Any("/vod/preView/:vodid/index.m3u8", vodHandler.Preview)
	router.Any("/sendfile/play/:file", sendfileHandler.Play)
	router.Any("/sendfile/down/:file", sendfileHandler.Down)
	router.Any("/comment/listing-:params", commentHandler.Listing)
	router.Any("/onego/rules", onegoHandler.Rules)
	router.Any("/onego/rooms", onegoHandler.Rooms)
	router.Any("/onego/current", onegoHandler.Current)
	router.Any("/onego/last", onegoHandler.Last)
	router.Any("/onego/hash", onegoHandler.Hash)
	for _, action := range []string{"listing", "recommend", "hot", "latest"} {
		router.Any("/vod/"+action, vodHandler.Listing)
		router.Any("/vod/"+action+"-:params", vodHandler.Listing)
	}

	v2 := router.Group("/v2")
	{
		v2.Any("/amazing/categories", amazingHandler.Categories)
		for _, action := range []string{"listing", "recommend", "hot", "latest"} {
			v2.Any("/amazing/"+action, amazingHandler.Listing)
			v2.Any("/amazing/"+action+"-:params", amazingHandler.Listing)
		}
		v2.Any("/so/list", soHandler.List)
		for _, action := range []string{"listing", "recommend", "hot", "latest"} {
			v2.Any("/vod/"+action, vodHandler.Listing)
			v2.Any("/vod/"+action+"-:params", vodHandler.Listing)
		}
		v2.Any("/register", notImplemented("c.apiv2.user.register"))
		v2.Any("/login", notImplemented("c.apiv2.user.login"))
		v2.Any("/forgot", notImplemented("c.apiv2.user.forgot"))
		v2.Any("/vod/show/:vodid", notImplemented("c.apiv2.vod.show"))
		for _, action := range []string{"up", "down", "reqplay", "reqdown", "buy"} {
			v2.Any("/vod/"+action+"/:vodid", notImplemented("c.apiv2.vod."+action))
		}
	}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not_found",
			"path":  c.Request.URL.Path,
		})
	})

	return router
}

type emptyIPLocator struct{}

func (emptyIPLocator) Find(string) ([]string, error) {
	return nil, nil
}

func newIPLocator(path string, logger *slog.Logger) iplocService.Locator {
	locator, err := iplocService.NewIPDBLocator(path)
	if err != nil {
		logger.Warn("ipdb unavailable, ip location responses will be empty", "path", path, "error", err)
		return emptyIPLocator{}
	}
	return locator
}

func openMySQL(dsn string, logger *slog.Logger) *sql.DB {
	if dsn == "" {
		return nil
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logger.Warn("mysql unavailable", "error", err)
		return nil
	}
	return db
}

func healthHandler(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, healthResponse{
			Status:  "ok",
			Service: "xj-comp-api",
			Env:     cfg.Env,
			Time:    time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func notImplemented(legacyHandler string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":          "not_implemented",
			"legacy_handler": legacyHandler,
		})
	}
}

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		logger.Info(
			"http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(started).Milliseconds(),
		)
	}
}
