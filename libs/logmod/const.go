package logmod

const (
	LogBidReq  = "bidrequest" // rtb 请求日志
	LogBidWin  = "bidwin"     // rtb 竞价成功日志
	LogBidImp  = "bidimpress" //rtb 展示日志
	LogBidClk  = "bidclick"   // rtb 点击日志
	LogBidConv = "bidconv"    //rtb 转化日志

	LogSspReq               = "ssprequest"           // ssp 请求日志
	LogSspImp               = "sspimpress"           // ssp 展示日志
	LogSspClk               = "sspclick"             // ssp 点击日志
	LogSspEvent             = "sspevent"             // ssp 事件日志
	LogSspConv              = "sspconv"              // ssp 转化日志
	LogInteractiveAdInImp   = "interact_ad_in_imp"   // 互动广告入口图片展示日志
	LogInteractiveAdInClick = "interact_ad_in_click" // 互动广告入口点击日志
	LogInteractiveAdPlay    = "interact_ad_play"     // 互动广告游戏点击日志
	LogInteractiveAdAccess  = "interact_ad_access"   // 互动广告访问日志
	LogInteractiveGetGames  = "interact_games_get"   // 互动广告获取游戏日志

	LogInitAdv         = "initadv"         //广告组初始化日志
	LogInitPub         = "initpub"         //媒体广告位初始化日志
	LogInitPolicy      = "initpolicy"      //策略定向初始化日志
	LogSystem          = "system"          //系统及日志
	LogInitNotice      = "initnotice"      //广告主回调通知地址初始化
	LogInitInteractive = "initinteractive" // 互动广告sspinfo 初始化
	LogInitRegion      = "initregion"      //省实信息初始化

	LogValidate = "logvalidate"
)
