package domain

type CaptchaReqData struct {
	PicURL     string `json:"picurl"`
	SMSCaptcha int    `json:"smscaptcha"`
}
