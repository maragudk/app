package model

type JobName string

func (j JobName) String() string {
	return string(j)
}

const (
	JobNameSendEmail JobName = "send-email"
)

type SendEmailJobData struct {
	Type     string
	Name     string
	Email    EmailAddress
	Keywords Keywords
}
