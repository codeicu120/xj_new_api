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
	activityRepo "xj_comp/internal/repository/activity"
	aiundressRepo "xj_comp/internal/repository/aiundress"
	amazingRepo "xj_comp/internal/repository/amazing"
	artRepo "xj_comp/internal/repository/art"
	boughtRepo "xj_comp/internal/repository/bought"
	commentRepo "xj_comp/internal/repository/comment"
	communityRepo "xj_comp/internal/repository/community"
	exploreRepo "xj_comp/internal/repository/explore"
	favoriteRepo "xj_comp/internal/repository/favorite"
	gameRepo "xj_comp/internal/repository/game"
	hgameRepo "xj_comp/internal/repository/hgame"
	historyRepo "xj_comp/internal/repository/history"
	indexRepo "xj_comp/internal/repository/index"
	inviteRepo "xj_comp/internal/repository/invite"
	minivodRepo "xj_comp/internal/repository/minivod"
	onegoRepo "xj_comp/internal/repository/onego"
	soRepo "xj_comp/internal/repository/so"
	statsRepo "xj_comp/internal/repository/stats"
	ucpRepo "xj_comp/internal/repository/ucp"
	userRepo "xj_comp/internal/repository/user"
	vodRepo "xj_comp/internal/repository/vod"
	activityService "xj_comp/internal/service/activity"
	aiundressService "xj_comp/internal/service/aiundress"
	amazingService "xj_comp/internal/service/amazing"
	artService "xj_comp/internal/service/art"
	attachService "xj_comp/internal/service/attach"
	boughtService "xj_comp/internal/service/bought"
	captchaService "xj_comp/internal/service/captcha"
	commentService "xj_comp/internal/service/comment"
	communityService "xj_comp/internal/service/community"
	exploreService "xj_comp/internal/service/explore"
	favoriteService "xj_comp/internal/service/favorite"
	gameService "xj_comp/internal/service/game"
	hgameService "xj_comp/internal/service/hgame"
	historyService "xj_comp/internal/service/history"
	indexService "xj_comp/internal/service/index"
	inviteService "xj_comp/internal/service/invite"
	iplocService "xj_comp/internal/service/iploc"
	minivodService "xj_comp/internal/service/minivod"
	onegoService "xj_comp/internal/service/onego"
	openService "xj_comp/internal/service/open"
	picService "xj_comp/internal/service/pic"
	sendfileService "xj_comp/internal/service/sendfile"
	soService "xj_comp/internal/service/so"
	statsService "xj_comp/internal/service/stats"
	ucpService "xj_comp/internal/service/ucp"
	userService "xj_comp/internal/service/user"
	verificationService "xj_comp/internal/service/verification"
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

	db := openMySQL(cfg.MySQLDSN, logger)
	userRepository := userRepo.NewRepository(db)
	userHandler := handler.NewUserHandler(userService.NewSysAvatarService(cfg.ResourceBaseURL), userService.NewLogoutService(userRepository))
	captchaHandler := handler.NewCaptchaHandler(captchaService.NewService(cfg.SMSCaptcha, cfg.CaptchaStyle, nil))
	testHandler := handler.NewTestHandler(captchaService.NewTestImageService())
	ipLocHandler := handler.NewIPLocHandler(iplocService.NewService(newIPLocator(cfg.IPDBPath, logger)))
	gameHandler := handler.NewGameHandler(
		gameService.NewPlatformService(gameRepo.NewPlatformRepository(db)),
		gameService.NewCategoryService(gameRepo.NewCategoryRepository(db), cfg.GameResourceURL),
		gameService.NewListingService(gameRepo.NewGameRepository(db), cfg.ResourceBaseURL),
		gameService.NewBroadcastService(gameRepo.NewBroadcastRepository(db)),
		gameService.NewWaliService(gameRepo.NewPlatformRepository(db), userRepository, nil),
	)
	indexRepository := indexRepo.NewSettingsRepository(db)
	ucpRepository := ucpRepo.NewRepository(db)
	vodRepository := vodRepo.NewListingRepository(db)
	vodListingService := vodService.NewListingService(vodRepository, cfg.ResourceBaseURL, cfg.VIPDiscount).WithAuth(userRepository)
	idxStore := indexStore{user: userRepository, ucp: ucpRepository, index: indexRepository}
	globalService := indexService.NewGlobalService(indexRepository, cfg.ResourceBaseURL)
	indexHandler := handler.NewIndexHandler(
		indexService.NewCertService(indexRepository, nil),
		globalService,
		indexService.NewInitService(idxStore, globalService, cfg.ResourceBaseURL),
		indexService.NewHomeService(vodRepository, vodListingService, cfg.ResourceBaseURL),
		indexService.NewCoverService(idxStore, nil, nil),
	)
	amazingHandler := handler.NewAmazingHandler(
		amazingService.NewCategoryService(amazingRepo.NewCategoryRepository(db)),
		amazingService.NewListingService(amazingRepo.NewSoftwareRepository(db), cfg.ResourceBaseURL),
	)
	soHandler := handler.NewSOHandler(soService.NewConfigService(soRepo.NewConfigRepository(db)))
	vodHandler := handler.NewVODHandler(vodListingService)
	minivodHandler := handler.NewMiniVODHandler(minivodService.NewService(minivodRepo.NewRepository(db), vodListingService, cfg.ResourceBaseURL).WithAuth(userRepository))
	specialHandler := handler.NewSpecialHandler(vodService.NewSpecialService(vodRepository, userRepository, cfg.ResourceBaseURL, 100))
	ucpHandler := handler.NewUCPHandler(ucpService.NewService(ucpStore{user: userRepository, ucp: ucpRepository, index: indexRepository}, cfg.ResourceBaseURL))
	sendfileHandler := handler.NewSendfileHandler(sendfileService.NewService(userRepository, vodRepository))
	commentHandler := handler.NewCommentHandler(commentService.NewService(commentRepo.NewRepository(db), cfg.ResourceBaseURL, userRepository))
	communityHandler := handler.NewCommunityHandler(communityService.NewService(userRepository, communityRepo.NewRepository(db), cfg.ResourceBaseURL))
	onegoHandler := handler.NewOneGoHandler(onegoService.NewService(onegoRepo.NewRepository(db), userRepository))
	artHandler := handler.NewArtHandler(artService.NewService(artRepo.NewRepository(db), cfg.ResourceBaseURL))
	attachHandler := handler.NewAttachHandler(attachService.NewService(userRepository))
	picHandler := handler.NewPicHandler(picService.NewService(cfg.UploadPath))
	statsRepository := statsRepo.NewRepository(db)
	statsHandler := handler.NewStatsHandler(statsService.NewService(statsRepository, userRepository))
	openHandler := handler.NewOpenHandler(openService.NewService(userRepository, statsRepository, cfg.ResourceBaseURL))
	activityHandler := handler.NewActivityHandler(activityService.NewService(activityRepo.NewRepository(db), userRepository, cfg.ResourceBaseURL))
	inviteHandler := handler.NewInviteHandler(inviteService.NewService(userRepository, inviteRepo.NewRepository(db)))
	boughtHandler := handler.NewBoughtHandler(boughtService.NewService(userRepository, boughtRepo.NewRepository(db), vodListingService))
	historyHandler := handler.NewHistoryHandler(historyService.NewService(userRepository, historyRepo.NewRepository(db), vodListingService))
	favoriteHandler := handler.NewFavoriteHandler(favoriteService.NewService(userRepository, favoriteRepo.NewRepository(db), vodListingService))
	exploreHandler := handler.NewExploreHandler(exploreService.NewService(userRepository, exploreRepo.NewRepository(db), cfg.ResourceBaseURL))
	hgameHandler := handler.NewHGameHandler(hgameService.NewService(hgameRepo.NewRepository(db), cfg.ResourceBaseURL))
	aiundressHandler := handler.NewAIUndressHandler(aiundressService.NewService(userRepository, aiundressRepo.NewRepository(db), cfg.ResourceBaseURL, cfg.Env))
	verificationHandler := handler.NewVerificationHandler(verificationService.NewService(idxStore, nil, nil, nil, nil))

	router.GET("/healthz", healthHandler(cfg))
	router.GET("/readyz", healthHandler(cfg))
	router.Any("/", indexHandler.Index)
	router.Any("/index", indexHandler.Index)
	router.Any("/sysavatar", userHandler.SysAvatar)
	router.Any("/logout", userHandler.Logout)
	router.Any("/sms", handler.EmptyHTML)
	router.Any("/sms/index", handler.EmptyHTML)
	router.Any("/sms/sendv", verificationHandler.SMSSendV)
	router.Any("/sms/sendu", verificationHandler.SMSSendU)
	router.Any("/email", handler.EmptyHTML)
	router.Any("/email/index", handler.EmptyHTML)
	router.Any("/email/send", verificationHandler.EmailSend)
	router.Any("/captcha/req", captchaHandler.Req)
	router.Any("/captcha/pic", captchaHandler.Pic)
	router.Any("/captcha/picx", captchaHandler.PicX)
	router.Any("/test", testHandler.Test)
	router.Any("/iploc/:ip", ipLocHandler.Find)
	router.Any("/game/platforms", gameHandler.Platforms)
	router.Any("/game/categories", gameHandler.Categories)
	router.Any("/game/games", gameHandler.Games)
	router.Any("/game/broadcasts", gameHandler.Broadcasts)
	router.Any("/game/wali/gameList", gameHandler.WaliGames)
	router.Any("/game/wali/test", gameHandler.WaliTest)
	router.Any("/game/wali/balance", gameHandler.WaliBalance)
	router.Any("/hgame/index", hgameHandler.Index)
	router.Any("/art", artHandler.Index)
	router.Any("/art/index", artHandler.Index)
	router.Any("/art/announce", artHandler.Announce)
	router.Any("/art/show", artHandler.Show)
	router.Any("/attach", attachHandler.Index)
	router.Any("/attach/index", attachHandler.Index)
	router.Any("/attach/upavatar", attachHandler.UpAvatar)
	router.Any("/getLikeRows", vodHandler.LikeRows)
	router.Any("/getCover", indexHandler.GetCover)
	router.Any("/getGlobalData", indexHandler.GetGlobalData)
	router.Any("/init", indexHandler.Init)
	router.Any("/search", vodHandler.Search)
	router.Any("/minisearch", vodHandler.MiniSearch)
	router.Any("/shortcutstats/add", statsHandler.ShortcutAdd)
	router.Any("/adstats/add", statsHandler.AdAdd)
	router.Any("/playstats/add", statsHandler.PlayAdd)
	router.Any("/open", openHandler.Index)
	router.Any("/open/index", openHandler.Index)
	router.Any("/open/reqauth", openHandler.ReqAuth)
	router.Any("/activity", activityHandler.Index)
	router.Any("/activity/index", activityHandler.Index)
	router.Any("/activity/details", activityHandler.Details)
	router.Any("/activity/luckyprizes", activityHandler.LuckyPrizes)
	router.Any("/activity/newyear2020", activityHandler.NewYear2020)
	router.Any("/activity/luckydraw", activityHandler.LuckyDraw)
	router.Any("/activity/luckydrawhistory", activityHandler.LuckyDrawHistory)
	router.Any("/activity/ranking", activityHandler.Ranking)
	router.Any("/activity/receive", activityHandler.Receive)
	router.Any("/activity/recommends", activityHandler.Recommends)
	router.Any("/invite/info", inviteHandler.Info)
	router.Any("/bought/listing", boughtHandler.Listing)
	router.Any("/bought/delete", boughtHandler.Delete)
	router.Any("/playlog", handler.EmptyHTML)
	router.Any("/playlog/index", handler.EmptyHTML)
	router.Any("/playlog/listing", historyHandler.PlayListing)
	router.Any("/playlog/remove", historyHandler.PlayRemove)
	router.Any("/downlog", handler.EmptyHTML)
	router.Any("/downlog/index", handler.EmptyHTML)
	router.Any("/downlog/listing", historyHandler.DownListing)
	router.Any("/downlog/remove", historyHandler.DownRemove)
	router.Any("/miniplaylog/listing", historyHandler.MiniPlayListing)
	router.Any("/miniplaylog/remove", historyHandler.MiniPlayRemove)
	router.Any("/favorite", handler.EmptyHTML)
	router.Any("/favorite/index", handler.EmptyHTML)
	router.Any("/favorite/listing", favoriteHandler.Listing)
	router.Any("/favorite/add", favoriteHandler.Add)
	router.Any("/favorite/remove", favoriteHandler.Remove)
	router.Any("/minifavorite", handler.EmptyHTML)
	router.Any("/minifavorite/index", handler.EmptyHTML)
	router.Any("/minifavorite/listing", favoriteHandler.MiniListing)
	router.Any("/minifavorite/add", favoriteHandler.MiniAdd)
	router.Any("/minifavorite/remove", favoriteHandler.MiniRemove)
	router.Any("/explore/index", exploreHandler.Index)
	router.Any("/explore/notification", exploreHandler.EmptyOK)
	router.Any("/explore/notification/index", exploreHandler.EmptyOK)
	router.Any("/explore/notification/clean", exploreHandler.CleanNotification)
	router.Any("/explore/signtask", exploreHandler.EmptyOK)
	router.Any("/explore/signtask/index", exploreHandler.EmptyOK)
	router.Any("/explore/vodtask", exploreHandler.EmptyOK)
	router.Any("/explore/vodtask/index", exploreHandler.EmptyOK)
	router.Any("/explore/vodtask/show/:vid", exploreHandler.VodTaskShow)
	router.Any("/aiundress", aiundressHandler.Listing)
	router.Any("/aiundress/listing", aiundressHandler.Listing)
	router.Any("/aiundress/index", handler.EmptyHTML)
	router.Any("/getCertUuid", indexHandler.GetCertUUID)
	router.Any("/ucp/index", ucpHandler.Index)
	router.Any("/ucp/user", ucpHandler.UserIndex)
	router.Any("/ucp/user/index", ucpHandler.UserIndex)
	router.Any("/ucp/bankcard", ucpHandler.BankcardIndex)
	router.Any("/ucp/bankcard/index", ucpHandler.BankcardIndex)
	router.Any("/ucp/bankcard/create", ucpHandler.BankcardCreate)
	router.Any("/ucp/bankcard/modify", ucpHandler.BankcardModify)
	router.Any("/ucp/bankcard/delete", ucpHandler.BankcardDelete)
	router.Any("/ucp/feedback", ucpHandler.FeedbackListing)
	router.GET("/ucp/feedback/index", ucpHandler.FeedbackIndex)
	router.GET("/ucp/feedback/listing", ucpHandler.FeedbackNewListing)
	router.GET("/ucp/feedback/detail", ucpHandler.FeedbackDetail)
	router.Any("/ucp/feedback/create", ucpHandler.FeedbackCreate)
	router.GET("/ucp/msg", ucpHandler.MsgListing)
	router.GET("/ucp/msg/index", ucpHandler.MsgListing)
	router.Any("/ucp/msg/show", ucpHandler.MsgDetail)
	router.Any("/ucp/msg/send", ucpHandler.MsgSend)
	router.Any("/ucp/msg/setread", ucpHandler.MsgSetRead)
	router.Any("/ucp/msg/cleanread", ucpHandler.MsgCleanRead)
	router.Any("/ucp/msg/delete", ucpHandler.MsgDelete)
	router.Any("/ucp/myaff", ucpHandler.MyAff)
	router.Any("/ucp/rolltitle", ucpHandler.RollTitle)
	router.Any("/ucp/task/sharepic", ucpHandler.TaskSharePic)
	router.Any("/ucp/task/qrlink", ucpHandler.TaskQRLink)
	router.Any("/ucp/taskbox/index", ucpHandler.TaskboxIndex)
	router.Any("/ucp/taskbox/taskboxlog", ucpHandler.TaskboxLog)
	router.Any("/ucp/taskbox/qrlink", ucpHandler.TaskboxQRLink)
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
	router.Any("/vod/up/:vodid", vodHandler.Up)
	router.Any("/vod/down/:vodid", vodHandler.Down)
	router.Any("/vod/preView/:vodid/index.m3u8", vodHandler.Preview)
	router.Any("/sendfile/play/:file", sendfileHandler.Play)
	router.Any("/sendfile/down/:file", sendfileHandler.Down)
	router.Any("/comment/listing-:params", commentHandler.Listing)
	router.Any("/comment/up", commentHandler.Up)
	router.Any("/comment/down", commentHandler.Down)
	for _, action := range []string{"list", "recommend", "hot", "latest", "favorite"} {
		router.Any("/community/"+action, communityHandler.Listing)
		router.Any("/community/"+action+"-:params", communityHandler.Listing)
	}
	router.Any("/community/clisting", communityHandler.CommentListing)
	router.Any("/community/clisting-:params", communityHandler.CommentListing)
	router.Any("/special/index", specialHandler.Index)
	router.Any("/special/listing", specialHandler.Listing)
	router.Any("/special/listing-:params", specialHandler.Listing)
	router.Any("/special/detail/:spid", specialHandler.Detail)
	router.Any("/special/up/:spid", specialHandler.Up)
	router.Any("/special/down/:spid", specialHandler.Down)
	router.Any("/onego/rules", onegoHandler.Rules)
	router.Any("/onego/rooms", onegoHandler.Rooms)
	router.Any("/onego/current", onegoHandler.Current)
	router.Any("/onego/last", onegoHandler.Last)
	router.Any("/onego/hash", onegoHandler.Hash)
	router.Any("/onego/history", onegoHandler.History)
	router.Any("/onego/lucky", onegoHandler.Lucky)
	router.Any("/onego/bet_ranks", onegoHandler.BetRanks)
	router.Any("/onego/marquee", onegoHandler.Marquee)
	for _, action := range []string{"listing", "recommend", "hot", "latest"} {
		router.Any("/vod/"+action, vodHandler.Listing)
		router.Any("/vod/"+action+"-:params", vodHandler.Listing)
	}
	for _, action := range []string{"listing", "recommend", "hot", "latest", "topzan", "topcomment", "topplay", "topcoin", "topnew", "topday", "topweek", "topmonth"} {
		router.Any("/minivod/"+action, minivodHandler.Listing)
		router.Any("/minivod/"+action+"-:params", minivodHandler.Listing)
	}
	router.Any("/minivod/show/:vodid", minivodHandler.Show)
	router.Any("/minivod/up/:vodid", minivodHandler.Up)
	router.Any("/minivod/down/:vodid", minivodHandler.Down)
	router.Any("/minivod/reqlong/:vodid", minivodHandler.ReqLong)
	router.Any("/my/:authorid", minivodHandler.Author)
	router.Any("/my/:authorid/:action", minivodHandler.Author)
	for _, size := range []string{"C1", "C2", "C3", "C4", "C5", "C6", "C7", "C8", "C9", "T1", "T2", "T3", "T4", "T5", "T6", "T7", "T8", "T9", "R1", "R2", "R3", "R4", "R5", "R6", "R7", "R8", "R9", "M", "N"} {
		router.Any("/"+size+"/*uri", picHandler.Index)
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
		v2.Any("/vod/show/:vodid", vodHandler.Show)
		v2.Any("/vod/up/:vodid", vodHandler.Up)
		v2.Any("/vod/down/:vodid", vodHandler.Down)
		for _, action := range []string{"reqplay", "reqdown", "buy"} {
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
