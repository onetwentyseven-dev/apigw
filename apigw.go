package apigw

import "github.com/sirupsen/logrus"

type Service struct {
	logger *logrus.Logger
}

func New(lgr *logrus.Logger) *Service {
	return &Service{
		logger: lgr,
	}
}
