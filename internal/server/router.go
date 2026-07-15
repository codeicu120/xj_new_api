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
	starliveRepo "xj_comp/internal/repository/starlive"
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
	paymentService "xj_comp/internal/service/payment"
	picService "xj_comp/internal/service/pic"
	respondService "xj_comp/internal/service/respond"
	sendfileService "xj_comp/internal/service/sendfile"
	soService "xj_comp/internal/service/so"
	starliveService "xj_comp/internal/service/starlive"
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
	userHandler := handler.NewUserHandler(
		userService.NewSysAvatarService(cfg.ResourceBaseURL),
		userService.NewLogoutService(userRepository),
		userService.NewAuthEdgeService(userRepository),
	)
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
	boughtHandler := handler.NewBoughtHandler(boughtService.NewService(userRepository, boughtRepo.NewRepository(db), vodListingService).WithVIPDiscount(cfg.VIPDiscount))
	historyHandler := handler.NewHistoryHandler(historyService.NewService(userRepository, historyRepo.NewRepository(db), vodListingService))
	favoriteHandler := handler.NewFavoriteHandler(favoriteService.NewService(userRepository, favoriteRepo.NewRepository(db), vodListingService).WithResourceBaseURL(cfg.ResourceBaseURL))
	exploreHandler := handler.NewExploreHandler(exploreService.NewService(userRepository, exploreRepo.NewRepository(db), cfg.ResourceBaseURL))
	hgameHandler := handler.NewHGameHandler(hgameService.NewService(hgameRepo.NewRepository(db), cfg.ResourceBaseURL))
	starLiveHandler := handler.NewStarLiveHandler(starliveService.NewService(starliveRepo.NewRepository(db), userRepository, ucpRepository, ucpRepository))
	aiundressExternalClient := aiundressService.NewHTTPExternalClient(cfg.AIUndressHost, cfg.AIUndressKey, 5*time.Second)
	aiundressHandler := handler.NewAIUndressHandler(aiundressService.NewService(userRepository, aiundressRepo.NewRepository(db), cfg.ResourceBaseURL, cfg.Env).WithExternalClient(aiundressExternalClient))
	verificationHandler := handler.NewVerificationHandler(verificationService.NewService(idxStore, nil, nil, nil, nil))
	paymentHandler := handler.NewPaymentHandler(paymentService.NewService(ucpStore{user: userRepository, ucp: ucpRepository, index: indexRepository}))
	respondHandler := handler.NewRespondHandler(respondService.NewService(ucpStore{user: userRepository, ucp: ucpRepository, index: indexRepository}))

	router.GET("/healthz", healthHandler(cfg))
	router.GET("/readyz", healthHandler(cfg))
	router.Any("/", indexHandler.Index)
	router.Any("/index", indexHandler.Index)
	router.Any("/sysavatar", userHandler.SysAvatar)
	router.Any("/register", userHandler.Register)
	router.Any("/login", userHandler.Login)
	router.Any("/logout", userHandler.Logout)
	router.Any("/forgot", userHandler.Forgot)
	router.Any("/delete", userHandler.Delete)
	router.Any("/changePhone", userHandler.ChangePhone)
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
	router.Any("/game/wali/topup", gameHandler.TransferTopup("瓦力游戏上分成功分支暂未迁移"))
	router.Any("/game/wali/withdraw", gameHandler.TransferWithdraw("瓦力游戏下分成功分支暂未迁移"))
	router.Any("/game/wali/enter", gameHandler.HighRiskAction("瓦力游戏进入成功分支暂未迁移"))
	router.Any("/game/lottery/gameList", gameHandler.LotteryGames)
	router.Any("/game/lottery/topup", gameHandler.TransferTopup("彩票游戏上分成功分支暂未迁移"))
	router.Any("/game/lottery/withdraw", gameHandler.TransferWithdraw("彩票游戏下分成功分支暂未迁移"))
	router.Any("/game/lottery/enter", gameHandler.HighRiskAction("彩票游戏进入成功分支暂未迁移"))
	router.Any("/game/lottery/balance", gameHandler.HighRiskAction("彩票游戏余额成功分支暂未迁移"))
	router.Any("/hgame/index", hgameHandler.Index)
	router.Any("/starLive/index", starLiveHandler.Index)
	router.Any("/starLive/queryCoinBalance", starLiveHandler.QueryCoinBalance)
	router.Any("/starLive/gameBet", starLiveHandler.GameBet)
	router.Any("/starLive/gameWin", starLiveHandler.GameWin)
	router.Any("/starLive/translate", starLiveHandler.Translate)
	router.Any("/starLive/tryAgain", starLiveHandler.TryAgain)
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
	router.Any("/invite/bind", inviteHandler.Bind)
	router.Any("/payment/index", paymentHandler.Query)
	router.Any("/payment/query", paymentHandler.Query)
	router.Any("/payment/payways", paymentHandler.Payways)
	router.Any("/payment/chpayway", paymentHandler.ChPayway)
	router.Any("/payment/unpaid", paymentHandler.Unpaid)
	router.Any("/payment/reqpay", paymentHandler.ReqPay)
	router.Any("/payment/pay12req", paymentHandler.Pay12Req)
	router.Any("/payment/success", paymentHandler.Success)
	router.Any("/payment/failed", paymentHandler.Failed)
	router.Any("/payment/wappay1", paymentHandler.WapPay1)
	router.Any("/payment/wappay2", paymentHandler.WapPay2)
	router.Any("/payment/pay7submit", paymentHandler.Pay7Submit)
	router.Any("/payment/pay11", paymentHandler.Pay11)
	for _, action := range []string{
		"shangfu", "wappay3", "wappay4", "wappay4a", "wappay5",
		"hawpay", "easypay", "pay6",
	} {
		router.Any("/payment/"+action, paymentHandler.Success)
	}
	for _, action := range []string{
		"pay7", "pay8", "pay9", "pay10", "pay10a", "pay10b", "pay12",
		"gpay1", "gpay2",
		"newpay", "newpayff", "newpayxxx", "newpayqk", "newpayxyf", "newpaykf", "newpaypi", "newpaygs", "newpaylep",
		"newpayys", "newpayyswx", "newpayhw", "newpayhs", "newpaypx", "newpaypxwx", "newpay99", "newpayxy", "newpayjd",
		"newpaycr", "newpaylu", "newpayluwx", "newpaymyr", "newpaymyrz", "newpaylh", "newpaylai", "newpayxh", "newpayya",
		"newpayyh", "newpayhf", "newpaydd", "newpaykk", "newpayrq",
	} {
		router.Any("/payment/"+action, paymentHandler.SuccessHTML)
	}
	for _, action := range []string{
		"shangfu", "wappay1", "wappay2", "wappay3", "wappay4", "wappay4a", "wappay5",
		"hawpay", "easypay", "gpay1", "gpay2",
		"pay6", "pay7", "pay8", "pay9", "pay10", "pay10a", "pay10b", "pay11",
		"newpay", "newpay99", "newpaycr", "newpaydd", "newpayff", "newpaygs", "newpayhs", "newpayhw",
		"newpayjd", "newpaylai", "newpaylh", "newpaylu", "newpayluwx", "newpaymyr", "newpaymyrz",
		"newpaypi", "newpaypx", "newpaypxwx", "newpayqk", "newpayrq", "newpayxh", "newpayxxx",
		"newpayxy", "newpayxyf", "newpayya", "newpayyh", "newpayys", "newpayyswx",
	} {
		router.Any("/respond/"+action, respondHandler.Failed("failed"))
	}
	for _, action := range []string{"pay12"} {
		router.Any("/respond/"+action, respondHandler.Failed("Err"))
	}
	for _, action := range []string{"newpayhf", "newpaykf", "newpaykk", "newpaylep"} {
		router.Any("/respond/"+action, respondHandler.Failed("FAILED"))
	}
	router.Any("/respond/chan1", respondHandler.Chan1)
	router.Any("/bought/listing", boughtHandler.Listing)
	router.Any("/bought/delete", boughtHandler.Delete)
	router.Any("/bought/buy", boughtHandler.Buy)
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
	router.Any("/explore/vodtask/reqcoin", exploreHandler.VodTaskReqCoin)
	router.Any("/aiundress", aiundressHandler.Listing)
	router.Any("/aiundress/listing", aiundressHandler.Listing)
	router.Any("/aiundress/index", handler.EmptyHTML)
	router.Any("/aiundress/upload", aiundressHandler.Upload)
	router.Any("/aiundress/undress", aiundressHandler.Undress)
	router.Any("/aiundress/delete", aiundressHandler.Delete)
	router.Any("/aiundress/moduleList", aiundressHandler.ModuleList)
	router.Any("/aiundress/resourceTypeList", aiundressHandler.ResourceTypeList)
	router.Any("/aiundress/resourceList", aiundressHandler.ResourceList)
	router.Any("/getCertUuid", indexHandler.GetCertUUID)
	router.Any("/ucp/index", ucpHandler.Index)
	router.Any("/ucp/user", ucpHandler.UserIndex)
	router.Any("/ucp/user/index", ucpHandler.UserIndex)
	router.Any("/ucp/user/profile", ucpHandler.UserProfile)
	router.Any("/ucp/user/passwd", ucpHandler.UserPasswd)
	router.Any("/ucp/user/checkemail", ucpHandler.UserCheckEmail)
	router.Any("/ucp/user/sendemail", ucpHandler.UserSendEmail)
	router.Any("/ucp/user/verifyemail", ucpHandler.UserVerifyEmail)
	router.Any("/ucp/user/bindmobi", ucpHandler.UserBindMobi)
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
	router.Any("/ucp/task", ucpHandler.TaskIndex)
	router.Any("/ucp/task/index", ucpHandler.TaskIndex)
	router.Any("/ucp/task/sharepic", ucpHandler.TaskSharePic)
	router.Any("/ucp/task/qrlink", ucpHandler.TaskQRLink)
	router.Any("/ucp/task/invite", ucpHandler.TaskInvite)
	router.Any("/ucp/task/sign", ucpHandler.TaskSign)
	router.Any("/ucp/task/share", ucpHandler.HighRiskAction("分享奖励成功分支暂未迁移"))
	router.Any("/ucp/task/qrcode", ucpHandler.HighRiskAction("二维码图片生成成功分支暂未迁移"))
	router.Any("/ucp/task/qrcodeSave", ucpHandler.HighRiskAction("保存二维码奖励成功分支暂未迁移"))
	router.Any("/ucp/task/invitecodeInput", ucpHandler.TaskInviteCodeInput)
	router.Any("/ucp/task/adviewClick", ucpHandler.TaskAdviewClick)
	router.Any("/ucp/taskbox/index", ucpHandler.TaskboxIndex)
	router.Any("/ucp/taskbox/taskboxlog", ucpHandler.TaskboxLog)
	router.Any("/ucp/taskbox/share", ucpHandler.TaskboxShare)
	router.Any("/ucp/taskbox/qrlink", ucpHandler.TaskboxQRLink)
	router.Any("/ucp/taskbox/taskboxopen", ucpHandler.TaskboxOpen)
	router.Any("/ucp/taskbox/qrcode", ucpHandler.HighRiskAction("任务宝箱二维码图片生成成功分支暂未迁移"))
	router.Any("/ucp/affcenter", ucpHandler.AffCenter)
	router.Any("/ucp/upgrade", ucpHandler.Upgrade)
	router.Any("/ucp/payment", ucpHandler.PaymentListing)
	router.Any("/ucp/payment/index", ucpHandler.PaymentListing)
	router.Any("/ucp/payment/listing", ucpHandler.PaymentListing)
	router.Any("/ucp/payment/safepaylog", ucpHandler.SafePayLog)
	router.Any("/ucp/account", ucpHandler.AccountIndex)
	router.Any("/ucp/account/index", ucpHandler.AccountIndex)
	router.Any("/ucp/account/balancelog", ucpHandler.BalanceLog)
	router.Any("/ucp/withdraw", ucpHandler.WithdrawIndex)
	router.Any("/ucp/withdraw/index", ucpHandler.WithdrawIndex)
	router.Any("/ucp/withdraw/listing", ucpHandler.WithdrawListing)
	router.Any("/ucp/withdraw/rule", ucpHandler.WithdrawRule)
	router.Any("/ucp/withdraw/create", ucpHandler.WithdrawCreate)
	router.Any("/ucp/coinlog", ucpHandler.CoinLogIndex)
	router.Any("/ucp/coinlog/index", ucpHandler.CoinLogIndex)
	router.Any("/ucp/coinlog/bonuslog", ucpHandler.CoinLogBonusLog)
	router.Any("/ucp/coinlog/invitelog", ucpHandler.CoinLogInviteLog)
	router.Any("/ucp/coinlog/exchange", ucpHandler.CoinLogExchange)
	router.Any("/ucp/vippkg", ucpHandler.VIPPkgIndex)
	router.Any("/ucp/vippkg/index", ucpHandler.VIPPkgIndex)
	router.Any("/ucp/vippkg/placeorder", ucpHandler.VIPPkgPlaceOrder)
	router.Any("/ucp/vippkg/coinorder", ucpHandler.VIPPkgCoinOrder)
	router.Any("/ucp/coinpkg", ucpHandler.CoinPkgIndex)
	router.Any("/ucp/coinpkg/index", ucpHandler.CoinPkgIndex)
	router.Any("/ucp/coinpkg/placeorder", ucpHandler.CoinPkgPlaceOrder)
	router.Any("/ucp/beanpkg", ucpHandler.BeanPkgIndex)
	router.Any("/ucp/beanpkg/index", ucpHandler.BeanPkgIndex)
	router.Any("/ucp/beanpkg/placeorder", ucpHandler.BeanPkgPlaceOrder)
	router.Any("/ucp/beanpkg/coinorder", ucpHandler.BeanPkgCoinOrder)
	router.Any("/ucp/vodorder", ucpHandler.VODOrderIndex)
	router.Any("/ucp/vodorder/index", ucpHandler.VODOrderIndex)
	router.Any("/ucp/vodorder/myorders", ucpHandler.VODOrderMyOrders)
	router.Any("/ucp/vodorder/mysupports", ucpHandler.VODOrderMySupports)
	router.Any("/ucp/vodorder/historyorders", ucpHandler.VODOrderHistoryOrders)
	router.Any("/ucp/vodorder/create", ucpHandler.VODOrderCreate)
	router.Any("/ucp/vodorder/support", ucpHandler.VODOrderSupport)
	router.Any("/vod/show/:vodid", vodHandler.Show)
	router.Any("/vod/up/:vodid", vodHandler.Up)
	router.Any("/vod/down/:vodid", vodHandler.Down)
	router.Any("/vod/reqplay/:vodid", vodHandler.ReqPlay)
	router.Any("/vod/reqdown/:vodid", vodHandler.ReqDown)
	router.Any("/vod/buy/:vodid", boughtHandler.Buy)
	router.Any("/vod/breaking", vodHandler.Breaking)
	router.Any("/vod/errorreport", vodHandler.ErrorReport)
	router.Any("/vod/preView/:vodid/index.m3u8", vodHandler.Preview)
	router.Any("/sendfile/play/:file", sendfileHandler.Play)
	router.Any("/sendfile/down/:file", sendfileHandler.Down)
	router.Any("/comment", handler.EmptyHTML)
	router.Any("/comment/index", handler.EmptyHTML)
	router.Any("/comment/listing-:params", commentHandler.Listing)
	router.Any("/comment/post", commentHandler.Post)
	router.Any("/comment/up", commentHandler.Up)
	router.Any("/comment/down", commentHandler.Down)
	for _, action := range []string{"list", "recommend", "hot", "latest", "favorite"} {
		router.Any("/community/"+action, communityHandler.Listing)
		router.Any("/community/"+action+"-:params", communityHandler.Listing)
	}
	router.Any("/community/show", communityHandler.Show)
	router.Any("/community/categories", communityHandler.Categories)
	router.Any("/community/slides", communityHandler.Slides)
	router.Any("/community/search", communityHandler.Search)
	router.Any("/community/clisting", communityHandler.CommentListing)
	router.Any("/community/clisting-:params", communityHandler.CommentListing)
	router.Any("/community/attention", communityHandler.Attention)
	router.Any("/community/up", communityHandler.Up)
	router.Any("/community/up_comment", communityHandler.UpComment)
	router.Any("/community/comment", communityHandler.Comment)
	router.Any("/community/post", communityHandler.Post)
	router.Any("/special/index", specialHandler.Index)
	router.Any("/special/listing", specialHandler.Listing)
	router.Any("/special/listing-:params", specialHandler.Listing)
	router.Any("/special/detail/:spid", specialHandler.Detail)
	router.Any("/special/up/:spid", specialHandler.Up)
	router.Any("/special/down/:spid", specialHandler.Down)
	router.Any("/onego", onegoHandler.Rules)
	router.Any("/onego/index", handler.EmptyHTML)
	router.Any("/onego/rules", onegoHandler.Rules)
	router.Any("/onego/rooms", onegoHandler.Rooms)
	router.Any("/onego/current", onegoHandler.Current)
	router.Any("/onego/last", onegoHandler.Last)
	router.Any("/onego/hash", onegoHandler.Hash)
	router.Any("/onego/history", onegoHandler.History)
	router.Any("/onego/bet", onegoHandler.Bet)
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
	router.Any("/minivod/reqlist", minivodHandler.ReqList)
	router.Any("/minivod/show/:vodid", minivodHandler.Show)
	router.Any("/minivod/up/:vodid", minivodHandler.Up)
	router.Any("/minivod/down/:vodid", minivodHandler.Down)
	router.Any("/minivod/throwcoin/:vodid", minivodHandler.ThrowCoin)
	router.Any("/minivod/reqplay/:vodid", minivodHandler.ReqPlay)
	router.Any("/minivod/reqdown/:vodid", minivodHandler.ReqDown)
	router.Any("/minivod/reqcoin", minivodHandler.ReqCoin)
	router.Any("/minivod/reqlong/:vodid", minivodHandler.ReqLong)
	router.Any("/minivod/parselong/:vodid/index.m3u8", minivodHandler.ParseLongM3U8)
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
		v2.Any("/captcha/req", captchaHandler.ReqV2)
		v2.Any("/captcha/pic", captchaHandler.Pic)
		v2.Any("/captcha/picx", captchaHandler.PicX)
		v2.Any("/captcha/verify", captchaHandler.Verify)
		v2.Any("/captcha/test", testHandler.Test)
		v2.Any("/so/list", soHandler.List)
		for _, action := range []string{"listing", "recommend", "hot", "latest"} {
			v2.Any("/vod/"+action, vodHandler.Listing)
			v2.Any("/vod/"+action+"-:params", vodHandler.Listing)
		}
		v2.Any("/register", userHandler.Register)
		v2.Any("/login", userHandler.LoginV2)
		v2.Any("/forgot", userHandler.ForgotV2)
		v2.Any("/vod/show/:vodid", vodHandler.Show)
		v2.Any("/vod/up/:vodid", vodHandler.Up)
		v2.Any("/vod/down/:vodid", vodHandler.Down)
		v2.Any("/vod/reqplay/:vodid", vodHandler.ReqPlay)
		v2.Any("/vod/reqdown/:vodid", vodHandler.ReqDown)
		v2.Any("/vod/buy/:vodid", boughtHandler.Buy)
		v2.Any("/vod/errorreport", vodHandler.ErrorReport)
		v2.Any("/minifavorite", handler.EmptyHTML)
		v2.Any("/minifavorite/index", handler.EmptyHTML)
		v2.Any("/minifavorite/listing", favoriteHandler.MiniV2Listing)
		v2.Any("/minifavorite/add", favoriteHandler.MiniAdd)
		v2.Any("/minifavorite/remove", favoriteHandler.MiniRemove)
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
