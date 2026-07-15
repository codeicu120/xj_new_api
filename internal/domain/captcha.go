package domain

type CaptchaReqData struct {
	PicURL     string `json:"picurl"`
	SMSCaptcha int    `json:"smscaptcha"`
}

type CaptchaReqV2Data struct {
	PicURL     string `json:"picurl"`
	SMSCaptcha int    `json:"smscaptcha"`
	CaptchaKey string `json:"captcha_key"`
}
